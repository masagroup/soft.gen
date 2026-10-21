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

// breakLongLines breaks the code lines of text longer than width as prettier
// breaks a call or a signature that does not fit: the first parenthesized
// list with items of the line gets one item per line, indented one level
// more, each item broken again when it is still too long. Comments and lines
// without such a list are kept.
func breakLongLines(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = breakLine(line, width)
	}
	return strings.Join(lines, "\n")
}

// breakLine breaks one line (see breakLongLines).
func breakLine(line string, width int) string {
	if len(line) <= width {
		return line
	}
	content := strings.TrimLeft(line, " ")
	indentation := line[:len(line)-len(content)]
	if strings.HasPrefix(content, "//") {
		return line
	}
	open, closing, items := firstList(content)
	if items == nil {
		return line
	}
	inner := indentation + indentUnit
	var b strings.Builder
	b.WriteString(indentation + content[:open+1] + "\n")
	if len(items) == 1 {
		// a single condition is broken at its top-level logical operators,
		// each operand on its line ending with the operator
		if operands := splitLogical(items[0]); len(operands) > 1 {
			for _, operand := range operands {
				b.WriteString(breakLine(inner+operand, width) + "\n")
			}
			b.WriteString(indentation + content[closing:])
			return b.String()
		}
	}
	for i, item := range items {
		if i < len(items)-1 {
			item += ","
		}
		b.WriteString(breakLine(inner+item, width) + "\n")
	}
	b.WriteString(indentation + content[closing:])
	return b.String()
}

// splitLogical splits an expression on its top-level || operators, else on
// its top-level && operators: the operands but the last one end with their
// operator, as prettier breaks a logical expression. A single operand means
// no top-level operator.
func splitLogical(expression string) []string {
	for _, operator := range []string{" || ", " && "} {
		var operands []string
		depth, start := 0, 0
		var quote byte
		for i := 0; i < len(expression); i++ {
			c := expression[i]
			switch {
			case quote != 0:
				if c == '\\' {
					i++
				} else if c == quote {
					quote = 0
				}
			case c == '"' || c == '\'' || c == '`':
				quote = c
			case c == '(' || c == '[' || c == '{':
				depth++
			case c == ')' || c == ']' || c == '}':
				depth--
			case depth == 0 && strings.HasPrefix(expression[i:], operator):
				operands = append(operands, expression[start:i]+" "+strings.TrimSpace(operator))
				start = i + len(operator)
				i = start - 1
			}
		}
		if len(operands) > 0 {
			return append(operands, expression[start:])
		}
	}
	return []string{expression}
}

// firstList finds the first parenthesized list of content with items: the
// offsets of its parentheses and its items split on the top-level commas.
// Strings are skipped.
func firstList(content string) (open, closing int, items []string) {
	depth, start := 0, -1
	var quote byte
	for i := 0; i < len(content); i++ {
		c := content[i]
		switch {
		case quote != 0:
			if c == '\\' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'' || c == '`':
			quote = c
		case c == '(' || c == '[' || c == '{' || c == '<':
			if c == '(' && depth == 0 {
				start = i
			}
			depth++
		case c == ')' || c == ']' || c == '}' || (c == '>' && (i == 0 || content[i-1] != '=')):
			depth--
			if c == ')' && depth == 0 && start >= 0 {
				if inside := strings.TrimSpace(content[start+1 : i]); inside != "" {
					return start, i, splitTopLevel(inside)
				}
				start = -1
			}
		}
	}
	return 0, 0, nil
}

// splitTopLevel splits a list on its commas outside brackets and strings.
func splitTopLevel(list string) []string {
	var items []string
	depth, start := 0, 0
	var quote byte
	for i := 0; i < len(list); i++ {
		c := list[i]
		switch {
		case quote != 0:
			if c == '\\' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'' || c == '`':
			quote = c
		case c == '(' || c == '[' || c == '{' || c == '<':
			depth++
		case c == ')' || c == ']' || c == '}' || (c == '>' && (i == 0 || list[i-1] != '=')):
			depth--
		case c == ',' && depth == 0:
			items = append(items, strings.TrimSpace(list[start:i]))
			start = i + 1
		}
	}
	return append(items, strings.TrimSpace(list[start:]))
}

// lines joins lines with line breaks: the template function writing lines
// built in Go.
func lines(items []string) string { return strings.Join(items, "\n") }
