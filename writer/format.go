package writer

import (
	"bytes"
	"fmt"
	"go/format"
	"os/exec"
	"path"
	"strings"
)

// Formatter formats the content of a generated file.
type Formatter interface {
	// Format returns the formatted content of the file at path (a slash
	// separated path relative to the output root, used to know the file type).
	Format(path string, content []byte) ([]byte, error)
}

// FormatterFunc adapts a function to Formatter.
type FormatterFunc func(path string, content []byte) ([]byte, error)

// Format calls f.
func (f FormatterFunc) Format(path string, content []byte) ([]byte, error) { return f(path, content) }

// GoFormatter formats .go files with go/format (gofmt) in process; other files
// are returned unchanged. A gofmt-clean output makes the downstream "go fmt" a
// no-op and the comparison with committed generated code direct.
type GoFormatter struct{}

// Format applies gofmt to .go files.
func (GoFormatter) Format(p string, content []byte) ([]byte, error) {
	if path.Ext(p) != ".go" {
		return content, nil
	}
	formatted, err := format.Source(content)
	if err != nil {
		return nil, fmt.Errorf("gofmt: %w (the generated code does not parse)", err)
	}
	return formatted, nil
}

// ExecFormatter formats files with an external command fed on its standard
// input and read on its standard output, such as "prettier --stdin-filepath
// {path}". The placeholder {path} in the arguments is replaced by the file path.
// Only files whose extension is in Extensions are formatted (all when empty).
type ExecFormatter struct {
	Command    string
	Args       []string
	Extensions []string
}

// Format runs the command on the content.
func (f ExecFormatter) Format(p string, content []byte) ([]byte, error) {
	if len(f.Extensions) > 0 && !contains(f.Extensions, path.Ext(p)) {
		return content, nil
	}
	args := make([]string, len(f.Args))
	for i, arg := range f.Args {
		args[i] = strings.ReplaceAll(arg, "{path}", p)
	}
	cmd := exec.Command(f.Command, args...)
	cmd.Stdin = bytes.NewReader(content)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", f.Command, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// Available reports whether the command of the formatter is on the PATH.
func (f ExecFormatter) Available() bool {
	_, err := exec.LookPath(f.Command)
	return err == nil
}

// contains reports whether list holds item.
func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
