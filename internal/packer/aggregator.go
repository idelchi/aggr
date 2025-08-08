// Package packer packs a set of small files into a single text stream and
// unpacks that stream back to files. Each unpacked file ends with exactly
// one trailing newline. Packing and unpacking use a worker pool whose size
// is controlled by Aggregator.Parallel.
package packer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/idelchi/godyl/pkg/logger"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"

	"gitlab.garfield-labs.com/apps/aggr/internal/checkers"
	"gitlab.garfield-labs.com/apps/aggr/internal/tree"
)

// Prefixes defines every token that appears in the packed stream.
type Prefixes struct {
	// Marker is the common prefix of BEGIN and END markers.
	Marker string
	// Begin is the suffix that starts a file section.
	Begin string
	// End is the suffix that ends a file section.
	End string
	// Escape is the replacement for Marker inside file content.
	Escape string
}

// Aggregator converts between files on disk and a packed stream.
type Aggregator struct {
	// Prefixes are the markers used in the packed stream.
	Prefixes Prefixes
	// Logger is the destination for debug information.
	Logger *logger.Logger
	// Dry controls whether to perform any file operations.
	Dry bool
	// Parallel is the maximum number of worker goroutines.
	Parallel int
	// Root is the root directory for packing.
	Root string
	// StripPrefix defines the prefix to strip from file paths.
	StripPrefix string
}

// fileChunk carries one file’s data from the parser to a worker.
type fileChunk struct {
	path string
	data []byte
}

// filesSink collects output file paths safely across workers.
type filesSink struct {
	mu sync.Mutex
	fs files.Files
}

// add appends f to the sink in a thread-safe manner.
func (s *filesSink) add(f file.File) {
	s.mu.Lock()
	s.fs.AddFile(f)
	s.mu.Unlock()
}

// NewAggregator returns an Aggregator with default markers and concurrency.
// Values ≤ 0 for parallel default to 1.
func NewAggregator(l *logger.Logger, dry bool, parallel int, root, stripPrefix string) *Aggregator {
	if parallel < 1 {
		parallel = 1
	}
	m := "// === AGGR:"
	return &Aggregator{
		Prefixes: Prefixes{
			Marker: m,
			Begin:  "BEGIN:",
			End:    "END:",
			Escape: m[:len("// ===")] + "\\ " + m[len("// === "):], // insert "\" before space
		},
		Logger:      l,
		Dry:         dry,
		Parallel:    parallel,
		Root:        root,
		StripPrefix: stripPrefix,
	}
}

// Pack writes a packed representation of set to a writer.
func (a *Aggregator) Pack(set files.Files, w io.Writer) error {
	if !a.Dry {
		if err := a.packFiles(set, w); err != nil {
			return err
		}
	}
	return a.writeFooter(set, w)
}

// Unpack reads a packed stream from r and recreates files under dst.
// It returns the list of files that were (or would be) written.
func (a *Aggregator) Unpack(r file.File, dst string, chk checkers.Checkers) (files.Files, error) {
	var sink filesSink

	eg, ctx := errgroup.WithContext(context.Background())
	eg.SetLimit(a.Parallel)

	ch := make(chan fileChunk, a.Parallel*2)

	// Worker goroutines.
	for i := 0; i < a.Parallel; i++ {
		eg.Go(func() error {
			for c := range ch {
				if err := a.writeChunk(c, dst, chk, &sink); err != nil {
					return err
				}
			}
			return nil
		})
	}

	// Parser goroutine.
	if err := a.parseStream(ctx, r, ch); err != nil {
		close(ch)
		_ = eg.Wait()
		return nil, err
	}
	close(ch)

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return sink.fs, nil
}

// packFiles packs every file in set and writes each block to w in order.
func (a *Aggregator) packFiles(set files.Files, w io.Writer) error {
	eg, _ := errgroup.WithContext(context.Background())
	eg.SetLimit(a.Parallel)

	blocks := make([][]byte, len(set))

	for i, f := range set {
		i, f := i, f
		eg.Go(func() error {
			b, err := a.packFile(f)
			if err != nil {
				return err
			}
			blocks[i] = b
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	for _, b := range blocks {
		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// packFile returns the packed representation of a single file.
func (a *Aggregator) packFile(f file.File) ([]byte, error) {
	realPath := file.New(a.Root, f.Path())

	data, err := realPath.Read()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", realPath, err)
	}

	var buf bytes.Buffer

	if a.StripPrefix != "" {
		f = f.WithoutFolder(a.StripPrefix)
	}

	fmt.Fprintf(&buf, "%s %s\n", a.Prefixes.beginPrefix(), f.Path())
	buf.Write(a.escape(canonical(data)))
	fmt.Fprintf(&buf, "%s %s\n\n", a.Prefixes.endPrefix(), f.Path())
	return buf.Bytes(), nil
}

// writeFooter appends the tree and file count summary.
func (a *Aggregator) writeFooter(set files.Files, w io.Writer) error {
	if _, err := io.WriteString(w, "\ntree\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, tree.Generate(set).String()); err != nil {
		return err
	}
	_, err := io.WriteString(w, fmt.Sprintf("\n%d files\n", len(set)))
	return err
}

func (a *Aggregator) parseStream(ctx context.Context, r file.File, ch chan<- fileChunk) error {
	f, err := r.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	begin := a.Prefixes.beginPrefix()
	end := a.Prefixes.endPrefix()

	var (
		curPath string
		buf     bytes.Buffer
		inFile  bool
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := br.ReadString('\n') // returns line w/ '\n' or EOF
		if err != nil && err != io.EOF && line == "" {
			return err
		}

		switch {
		case strings.HasPrefix(line, begin):
			if inFile {
				return fmt.Errorf("nested %q for %s", begin, curPath)
			}
			curPath = strings.TrimSpace(line[len(begin):])
			buf.Reset()
			inFile = true

		case strings.HasPrefix(line, end):
			p := strings.TrimSpace(line[len(end):])
			if !inFile || p != curPath {
				return fmt.Errorf("%q without matching %q for %s", end, begin, p)
			}
			dataCopy := append([]byte(nil), buf.Bytes()...)
			ch <- fileChunk{path: curPath, data: dataCopy}
			inFile = false

		default:
			if inFile {
				buf.WriteString(line) // preserve newlines as before
			}
		}

		if err == io.EOF {
			break
		}
	}

	if inFile {
		return fmt.Errorf("unterminated file %q", curPath)
	}
	return nil
}

// writeChunk writes one unpacked file to disk unless Dry is true.
func (a *Aggregator) writeChunk(c fileChunk, dst string, chk checkers.Checkers, sink *filesSink) error {
	data := canonical(a.unescape(c.data))

	if err := chk.Check(c.path); err != nil {
		a.Logger.Debugf("  - %s: %v", c.path, err)
		return nil
	}

	f := file.New(dst, c.path)
	sink.add(f)

	if a.Dry {
		return nil
	}
	if err := f.Create(); err != nil {
		return fmt.Errorf("create %s: %w", f, err)
	}
	w, err := f.OpenForWriting()
	if err != nil {
		return fmt.Errorf("open %s: %w", f, err)
	}
	defer w.Close()

	_, err = w.Write(data)
	return err
}

// beginPrefix returns the full BEGIN marker used during parsing.
func (p Prefixes) beginPrefix() string { return fmt.Sprintf("%s %s", p.Marker, p.Begin) }

// endPrefix returns the full END marker used during parsing.
func (p Prefixes) endPrefix() string { return fmt.Sprintf("%s %s", p.Marker, p.End) }

func (a *Aggregator) escape(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), a.Prefixes.Marker) {
			lines[i] = strings.Replace(line, a.Prefixes.Marker, a.Prefixes.Escape, 1)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

func (a *Aggregator) unescape(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), a.Prefixes.Escape) {
			lines[i] = strings.Replace(line, a.Prefixes.Escape, a.Prefixes.Marker, 1)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// canonical trims all trailing newlines and appends one '\n'.
func canonical(data []byte) []byte {
	data = bytes.TrimRight(data, "\n")
	return append(data, '\n')
}
