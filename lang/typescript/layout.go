package typescript

import (
	"strings"
)

// The layout of the generated code is the one prettier gives it with the
// options of soft.ts: an indentation of 4 spaces, no semicolons, double
// quotes, no trailing commas and a print width of 120 (the printWidth
// property). The templates write the fixed layout; the lists the generator
// builds are wrapped here as prettier wraps them.

// indentUnit is one level of indentation.
const indentUnit = "    "

// indent indents the non-empty lines of s by n levels: the TypeScript
// version of the template function.
func indent(n int, s string) string {
	prefix := strings.Repeat(indentUnit, n)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// quote writes a string literal in double quotes, escaping the backslashes,
// the double quotes and the line breaks.
func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`).Replace(s) + `"`
}

// wrapList writes a list between an opening and a closing text, its items
// separated by ", " on one line when it fits in the width at that level of
// indentation, else one item per line, indented one level more, as prettier
// breaks import specifiers, parameters and arguments:
//
//	import { A, B } from "x"
//
//	import {
//	    A,
//	    B
//	} from "x"
func wrapList(open string, items []string, closing string, level, width int) string {
	space := ""
	if strings.HasSuffix(open, "{") {
		space = " "
	}
	line := open + space + strings.Join(items, ", ") + space + closing
	if len(items) == 0 {
		line = strings.TrimSuffix(open, " ") + closing
	}
	if len(items) == 0 || level*len(indentUnit)+len(line) <= width {
		return line
	}
	inner := strings.Repeat(indentUnit, level+1)
	var b strings.Builder
	b.WriteString(open + "\n")
	for i, item := range items {
		b.WriteString(inner + item)
		if i < len(items)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat(indentUnit, level) + closing)
	return b.String()
}
