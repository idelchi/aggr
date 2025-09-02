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

	"github.com/idelchi/aggr/internal/checkers"
	"github.com/idelchi/aggr/internal/tree"
	"github.com/idelchi/godyl/pkg/logger"
	"github.com/idelchi/godyl/pkg/path/file"
	"github.com/idelchi/godyl/pkg/path/files"
)

// Prefixes defines the markers and tokens used in the packed stream format.
type Prefixes struct {
	// Marker is the common prefix used for BEGIN and END markers.
	Marker string
	// Begin is the suffix that indicates the start of a file section.
	Begin string
	// End is the suffix that indicates the end of a file section.
	End string
	// Escape is the replacement string for Marker when it appears inside file content.
	Escape string
}

// Aggregator handles the conversion between individual files and packed streams.
// It supports both packing multiple files into a single stream and unpacking
// such streams back into individual files.
type Aggregator struct {
	// Prefixes contains the markers used in the packed stream format.
	Prefixes Prefixes
	// Logger receives debug and informational messages during processing.
	Logger *logger.Logger
	// Dry indicates whether to perform actual file operations or just simulate them.
	Dry bool
	// Parallel sets the maximum number of concurrent worker goroutines.
	Parallel int
	// Root specifies the root directory for file operations during packing.
	Root string
}

// fileChunk carries one file's data from the parser to a worker.
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

// NewAggregator creates a new Aggregator with default configuration.
// If parallel is ≤ 0, it defaults to 1 worker. The aggregator uses predefined
// markers for the packed stream format.
func NewAggregator(log *logger.Logger, dry bool, parallel int, root string) *Aggregator {
	if parallel < 1 {
		parallel = 1
	}

	marker := "// === AGGR:"

	return &Aggregator{
		Prefixes: Prefixes{
			Marker: marker,
			Begin:  "BEGIN:",
			End:    "END:",
			Escape: marker[:len("// ===")] + "\\ " + marker[len("// === "):], // insert "\" before space
		},
		Logger:   log,
		Dry:      dry,
		Parallel: parallel,
		Root:     root,
	}
}

// Pack writes a packed representation of the file set to the provided writer.
// It processes all files concurrently and writes them in the packed format.
func (a *Aggregator) Pack(set files.Files, writer io.Writer) error {
	if !a.Dry {
		if err := a.packFiles(set, writer); err != nil {
			return err
		}
	}

	return a.writeFooter(set, writer)
}

// Unpack reads a packed stream and recreates the original files under the destination directory.
// It returns the list of files that were written (or would be written in dry run mode).
// The checkers parameter allows filtering which files to extract.
func (a *Aggregator) Unpack(reader file.File, dst string, chk checkers.Checkers) (files.Files, error) {
	var sink filesSink

	errGroup, ctx := errgroup.WithContext(context.Background())
	errGroup.SetLimit(a.Parallel + 1)

	const channelBufferFactor = 2

	chunks := make(chan fileChunk, a.Parallel*channelBufferFactor)

	// Worker goroutines.
	for range a.Parallel {
		errGroup.Go(func() error {
			for chunk := range chunks {
				if err := a.writeChunk(chunk, dst, chk, &sink); err != nil {
					return err
				}
			}

			return nil
		})
	}

	// Parser goroutine.
	errGroup.Go(func() error {
		defer close(chunks)

		return a.parseStream(ctx, reader, chunks)
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return sink.fs, nil
}

// packFiles packs every file in set and writes each block to w in order.
func (a *Aggregator) packFiles(set files.Files, writer io.Writer) error {
	errGroup, _ := errgroup.WithContext(context.Background())
	errGroup.SetLimit(a.Parallel)

	blocks := make([][]byte, len(set))

	for index, file := range set {
		errGroup.Go(func() error {
			b, err := a.packFile(file)
			if err != nil {
				return err
			}

			blocks[index] = b

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}

	for _, b := range blocks {
		if _, err := writer.Write(b); err != nil {
			return err
		}
	}

	return nil
}

// packFile returns the packed representation of a single file.
func (a *Aggregator) packFile(inputFile file.File) ([]byte, error) {
	realPath := file.New(a.Root, inputFile.Path())

	data, err := realPath.Read()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", realPath, err)
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s %s\n", a.Prefixes.beginPrefix(), inputFile.Path())
	buf.Write(a.escape(canonical(data)))
	fmt.Fprintf(&buf, "%s %s\n\n", a.Prefixes.endPrefix(), inputFile.Path())

	return buf.Bytes(), nil
}

// writeFooter appends the tree and file count summary.
func (a *Aggregator) writeFooter(set files.Files, writer io.Writer) error {
	if _, err := io.WriteString(writer, "\ntree\n"); err != nil {
		return err
	}

	if _, err := io.WriteString(writer, tree.Generate(set, a.Dry).String()); err != nil {
		return err
	}

	_, err := io.WriteString(writer, fmt.Sprintf("\n%d files\n", len(set)))

	return err
}

// parseStream reads a packed stream and sends file chunks to the provided channel.
//
//nolint:gocognit	// Function is complex by design.
func (a *Aggregator) parseStream(ctx context.Context, reader file.File, chunks chan<- fileChunk) error {
	openFile, err := reader.Open()
	if err != nil {
		return err
	}
	defer openFile.Close()

	bufReader := bufio.NewReader(openFile)
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

		line, err := bufReader.ReadString('\n') // returns line w/ '\n' or EOF
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
			chunks <- fileChunk{path: curPath, data: dataCopy}

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
func (a *Aggregator) writeChunk(chunk fileChunk, dst string, checkers checkers.Checkers, sink *filesSink) error {
	data := canonical(a.unescape(chunk.data))

	if err := checkers.Check("", chunk.path); err != nil {
		a.Logger.Debugf("  - %s: %v", chunk.path, err)

		return nil
	}

	outputFile := file.New(dst, chunk.path)
	sink.add(outputFile)

	if a.Dry {
		return nil
	}

	if err := outputFile.Create(); err != nil {
		return fmt.Errorf("create %s: %w", outputFile, err)
	}

	fileWriter, err := outputFile.OpenForWriting()
	if err != nil {
		return fmt.Errorf("open %s: %w", outputFile, err)
	}
	defer fileWriter.Close()

	_, err = fileWriter.Write(data)

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
