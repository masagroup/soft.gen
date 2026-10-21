package writer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Status is what Write did with a file.
type Status int

const (
	// Written means the file was written (created or overwritten).
	Written Status = iota
	// Skipped means dry-run mode: the file was not written.
	Skipped
	// Same means diff mode: the existing file already has the generated content.
	Same
	// Different means diff mode: the existing file differs (or is missing); nothing was written.
	Different
)

// String names the status.
func (s Status) String() string {
	switch s {
	case Written:
		return "written"
	case Skipped:
		return "skipped"
	case Same:
		return "same"
	case Different:
		return "different"
	}
	return fmt.Sprintf("Status(%d)", int(s))
}

// Flags select the treatments a file gets.
type Flags uint8

const (
	// WithHeader prepends the generated-code header to the file.
	WithHeader Flags = 1 << iota
	// WithRegions preserves the user code regions of the existing file.
	WithRegions
)

// Writer writes generated files under a root directory.
type Writer struct {
	root      string
	silent    bool
	dryRun    bool
	diff      bool
	header    func(path string) string
	regions   *RegionSyntax
	formatter Formatter
	output    io.Writer
	errOutput io.Writer
}

// Option configures a Writer.
type Option func(*Writer)

// Silent turns the "file generated:" lines off.
func Silent(silent bool) Option { return func(w *Writer) { w.silent = silent } }

// DryRun logs the files without writing them.
func DryRun(dryRun bool) Option { return func(w *Writer) { w.dryRun = dryRun } }

// Diff prints a unified diff against the existing files and writes nothing.
func Diff(diff bool) Option { return func(w *Writer) { w.diff = diff } }

// Header sets the function giving the generated-code header of a path (the
// header is prepended to files written with WithHeader).
func Header(header func(path string) string) Option {
	return func(w *Writer) { w.header = header }
}

// Regions sets the syntax of the user code regions (nil: no regions).
func Regions(syntax *RegionSyntax) Option { return func(w *Writer) { w.regions = syntax } }

// Format sets the formatter applied to every file (nil: none).
func Format(formatter Formatter) Option { return func(w *Writer) { w.formatter = formatter } }

// Output sets where the "file generated:" lines and the diffs go (default: stdout).
func Output(out io.Writer) Option { return func(w *Writer) { w.output = out } }

// ErrOutput sets where the warnings go (default: stderr).
func ErrOutput(out io.Writer) Option { return func(w *Writer) { w.errOutput = out } }

// New returns a Writer writing under root.
func New(root string, opts ...Option) *Writer {
	w := &Writer{root: root, output: os.Stdout, errOutput: os.Stderr}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Root returns the output root directory.
func (w *Writer) Root() string { return w.root }

// Write writes the generated content of the file at path (slash-separated,
// relative to the root): the line endings are normalized to LF, a final
// newline is ensured, the header is prepended (WithHeader), the user code
// regions of the existing file are preserved (WithRegions), the content is
// formatted, the parent directories are created and the file is overwritten.
// In dry-run mode nothing is written; in diff mode the difference with the
// existing file is printed instead.
func (w *Writer) Write(path string, content []byte, flags Flags) (Status, error) {
	full := filepath.Join(w.root, filepath.FromSlash(path))
	existing, err := os.ReadFile(full)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Different, fmt.Errorf("%s: %w", path, err)
	}
	content = normalizeNewlines(content)
	if flags&WithHeader != 0 && w.header != nil {
		content = append([]byte(w.header(path)), content...)
	}
	if flags&WithRegions != 0 && w.regions != nil && exists {
		merged, lost, err := PreserveRegions(*w.regions, existing, content)
		if err != nil {
			return Different, fmt.Errorf("%s: %w", path, err)
		}
		for _, id := range lost {
			fmt.Fprintf(w.errOutput, "warning: %s: user code region %q is not generated any more, its content is lost\n", path, id)
		}
		content = merged
	}
	if w.formatter != nil {
		formatted, err := w.formatter.Format(path, content)
		if err != nil {
			return Different, fmt.Errorf("%s: format: %w", path, err)
		}
		content = formatted
	}
	switch {
	case w.diff:
		if exists && bytes.Equal(normalizeNewlines(existing), content) {
			return Same, nil
		}
		fmt.Fprint(w.output, UnifiedDiff(path, normalizeNewlines(existing), content))
		return Different, nil
	case w.dryRun:
		w.log(full)
		return Skipped, nil
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return Different, err
	}
	if err := os.WriteFile(full, content, 0o644); err != nil {
		return Different, err
	}
	w.log(full)
	return Written, nil
}

// log prints the "file generated:" line of v1 unless silent.
func (w *Writer) log(full string) {
	if !w.silent {
		fmt.Fprintf(w.output, "file generated: %s\n", filepath.Clean(full))
	}
}

// normalizeNewlines turns CRLF into LF and ensures a final newline.
func normalizeNewlines(content []byte) []byte {
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	return content
}
