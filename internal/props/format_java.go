package props

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// javaFormat parses Java .properties files as java.util.Properties.load does:
// key=value, key:value or key value lines; # and ! comments; leading
// whitespace ignored; a trailing backslash continues the line and the leading
// whitespace of the next line is dropped; the escapes \n \t \r \f \\ \uXXXX
// are resolved and a backslash before any other character yields that
// character.
type javaFormat struct{}

func init() { register(javaFormat{}) }

func (javaFormat) name() string         { return "properties" }
func (javaFormat) extensions() []string { return []string{".properties"} }

func (javaFormat) parse(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	values := map[string]string{}
	for i := 0; i < len(lines); i++ {
		line := trimLeadingSpace(lines[i])
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		// logical line: join the continuation lines
		for endsWithContinuation(line) {
			line = line[:len(line)-1]
			if i+1 >= len(lines) { // continuation at end of file
				break
			}
			i++
			line += trimLeadingSpace(lines[i])
		}
		key, value, err := splitJavaLine(line)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
}

// trimLeadingSpace drops the leading whitespace Java ignores.
func trimLeadingSpace(s string) string {
	return strings.TrimLeft(s, " \t\f")
}

// isJavaSpace reports whether c is whitespace for the Java properties syntax.
func isJavaSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\f'
}

// endsWithContinuation reports whether the line ends with an odd number of
// backslashes, which continues it on the next line.
func endsWithContinuation(line string) bool {
	n := 0
	for i := len(line) - 1; i >= 0 && line[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

// splitJavaLine splits a logical line into its unescaped key and value: the
// key ends at the first unescaped '=', ':' or whitespace; then whitespace, one
// optional separator and whitespace are skipped and the rest is the value.
func splitJavaLine(line string) (key, value string, err error) {
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if c == '\\' {
			i++ // skip the escaped character
			continue
		}
		if c == '=' || c == ':' || isJavaSpace(c) {
			break
		}
	}
	rawKey := line[:min(i, len(line))]
	for i < len(line) && isJavaSpace(line[i]) {
		i++
	}
	if i < len(line) && (line[i] == '=' || line[i] == ':') {
		i++
	}
	for i < len(line) && isJavaSpace(line[i]) {
		i++
	}
	if key, err = unescapeJava(rawKey); err != nil {
		return "", "", err
	}
	if value, err = unescapeJava(line[i:]); err != nil {
		return "", "", err
	}
	return key, value, nil
}

// unescapeJava resolves the backslash escapes of a key or value.
func unescapeJava(s string) (string, error) {
	if !strings.Contains(s, "\\") {
		return s, nil
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			sb.WriteByte(c)
			continue
		}
		i++
		if i >= len(s) { // lone trailing backslash
			break
		}
		switch c = s[i]; c {
		case 'n':
			sb.WriteByte('\n')
		case 't':
			sb.WriteByte('\t')
		case 'r':
			sb.WriteByte('\r')
		case 'f':
			sb.WriteByte('\f')
		case 'u':
			if i+4 >= len(s) {
				return "", fmt.Errorf("malformed \\uXXXX encoding in %q", s)
			}
			code, err := strconv.ParseUint(s[i+1:i+5], 16, 16)
			if err != nil {
				return "", fmt.Errorf("malformed \\uXXXX encoding in %q", s)
			}
			sb.WriteRune(rune(code))
			i += 4
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String(), nil
}
