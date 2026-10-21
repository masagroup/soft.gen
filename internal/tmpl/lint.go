package tmpl

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"
	"text/template/parse"
)

// Problem is a violation of the template conventions.
type Problem struct {
	File    string
	Line    int
	Message string
}

// String formats the problem as file:line: message.
func (p Problem) String() string {
	return fmt.Sprintf("%s:%d: %s", p.File, p.Line, p.Message)
}

var (
	defineLine     = regexp.MustCompile(`^\s*\{\{\s*define\s+"([^"]+)"\s*(-?)\}\}\s*$`)
	controlLine    = regexp.MustCompile(`^\s*\{\{(-?)\s*(if|else|range|with|end|break|continue)\b.*?(-?)\}\}\s*$`)
	newlineLiteral = regexp.MustCompile(`\{\{-?\s*"\\n"`)
)

// Lint checks the conventions of the template files of fsys:
//
//  1. a file holds only define blocks (and comments) at its top level;
//  2. a define is named after its file: a file dir/name.tmpl defines
//     "dir/name", "dir/name.<variant>" or "dir/name/<sub>";
//  3. a define opens with {{ define "name" -}} and closes with {{- end }};
//  4. a line holding only a control action (if, else, range, with, end, break,
//     continue) is written {{- action }}: left trim only;
//  5. no {{ "\n" }} newline literal;
//  6. no two consecutive blank lines.
//
// A template that passes Lint renders complete lines, owns no trailing
// newline, and leaves no trace of its control lines in the output. funcs are
// the functions a language adds to the base ones (Language.Funcs).
func Lint(fsys fs.FS, funcs ...template.FuncMap) []Problem {
	var problems []Problem
	paths, err := templatePaths(fsys)
	if err != nil {
		return []Problem{{Message: err.Error()}}
	}
	for _, path := range paths {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			problems = append(problems, Problem{File: path, Message: err.Error()})
			continue
		}
		problems = append(problems, lintFile(path, string(content), funcs)...)
	}
	return problems
}

// lintFile checks one template file.
func lintFile(path, content string, funcs []template.FuncMap) []Problem {
	var problems []Problem
	report := func(line int, format string, args ...any) {
		problems = append(problems, Problem{File: path, Line: line, Message: fmt.Sprintf(format, args...)})
	}
	prefix := strings.TrimSuffix(path, Extension)
	lines := strings.Split(content, "\n")
	blankRun := 0
	for i, line := range lines {
		n := i + 1
		if strings.TrimSpace(line) == "" {
			blankRun++
			if blankRun == 2 {
				report(n, "two consecutive blank lines")
			}
			continue
		}
		blankRun = 0
		if m := defineLine.FindStringSubmatch(line); m != nil {
			name := m[1]
			if name != prefix && !strings.HasPrefix(name, prefix+".") && !strings.HasPrefix(name, prefix+"/") {
				report(n, "define %q is not named after its file (%q or \"%s.<variant>\")", name, prefix, prefix)
			}
			if m[2] != "-" {
				report(n, "define %q must open with {{ define %q -}}", name, name)
			}
			continue
		}
		if m := controlLine.FindStringSubmatch(line); m != nil {
			if m[1] != "-" {
				report(n, "control line must be left-trimmed: {{- %s ... }}", m[2])
			}
			if m[3] == "-" {
				report(n, "control line must not be right-trimmed: {{- %s ... }}", m[2])
			}
		}
		if newlineLiteral.MatchString(line) {
			report(n, `newline literal {{ "\n" }}: use a blank line in the template`)
		}
	}
	parsed, err := template.New(path).Funcs(lintFuncs(funcs)).Parse(content)
	if err != nil {
		report(0, "parse: %v", err)
		return problems
	}
	for _, node := range parsed.Root.Nodes {
		switch node := node.(type) {
		case *parse.TextNode:
			if strings.TrimSpace(string(node.Text)) != "" {
				report(lineOf(content, node.Position()), "text outside a define block")
			}
		case *parse.CommentNode:
		default:
			report(lineOf(content, node.Position()), "action outside a define block: %s", node)
		}
	}
	return problems
}

// lintFuncs declares every function name, the base ones and the ones of
// the language, so that parsing succeeds.
func lintFuncs(extra []template.FuncMap) template.FuncMap {
	funcs := BaseFuncs()
	funcs["include"] = func(string, any) string { return "" }
	for _, more := range extra {
		for name, fn := range more {
			funcs[name] = fn
		}
	}
	return funcs
}

// lineOf returns the 1-based line of a byte offset.
func lineOf(content string, pos parse.Pos) int {
	return 1 + strings.Count(content[:int(pos)], "\n")
}
