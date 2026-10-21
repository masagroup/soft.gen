package naming

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Style is the syntax of the naming rules of a language.
type Style int

const (
	// Regex is the "pattern:replacement" (or bare "replacement") style of the
	// Go and TypeScript generators, with the semantics of the Java
	// String.replaceFirst: the pattern is a regular expression and the
	// replacement refers to its groups with $1 or ${1}.
	Regex Style = iota
	// MessageFormat is the "{0}Base" style of the C++ generator, where every
	// "{0}" placeholder stands for the whole name.
	MessageFormat
)

// String returns the name of the style, "Regex" or "MessageFormat".
func (s Style) String() string {
	switch s {
	case Regex:
		return "Regex"
	case MessageFormat:
		return "MessageFormat"
	default:
		return fmt.Sprintf("Style(%d)", int(s))
	}
}

// Rule is a parsed naming rule, ready to be applied to element names.
type Rule struct {
	style   Style
	pattern *regexp.Regexp // Regex only
	// template is the Go regexp.Expand template for Regex, or the format
	// string with its "{0}" placeholders for MessageFormat.
	template string
}

// Style returns the style the rule was parsed with.
func (r Rule) Style() Style {
	return r.style
}

// placeholder is the whole-name placeholder of the MessageFormat style.
const placeholder = "{0}"

// messageFormatArg matches any "{N}" argument of a MessageFormat rule.
var messageFormatArg = regexp.MustCompile(`\{(\d+)\}`)

// ParseRule parses a rule spec in the given style. In the Regex style, the
// spec is split on its first ':' into a regular expression pattern and a Java
// replacement, and a spec without ':' is a replacement applied to the pattern
// "(.*)"; "$1" and "${1}" both refer to the first group, and "\$" is a literal
// dollar. In the MessageFormat style, the spec is kept as is and may only
// contain "{0}" placeholders. An invalid pattern or replacement is an error.
func ParseRule(spec string, style Style) (Rule, error) {
	switch style {
	case Regex:
		pattern, replacement := "(.*)", spec
		if index := strings.IndexByte(spec, ':'); index >= 0 {
			pattern, replacement = spec[:index], spec[index+1:]
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return Rule{}, fmt.Errorf("naming: rule %q: %w", spec, err)
		}
		template, err := javaReplacementToTemplate(replacement, re.NumSubexp())
		if err != nil {
			return Rule{}, fmt.Errorf("naming: rule %q: %w", spec, err)
		}
		return Rule{style: Regex, pattern: re, template: template}, nil
	case MessageFormat:
		for _, arg := range messageFormatArg.FindAllStringSubmatch(spec, -1) {
			if arg[0] != placeholder {
				return Rule{}, fmt.Errorf("naming: rule %q: unknown placeholder %s, only %s is allowed", spec, arg[0], placeholder)
			}
		}
		return Rule{style: MessageFormat, template: spec}, nil
	default:
		return Rule{}, fmt.Errorf("naming: rule %q: unknown style %s", spec, style)
	}
}

// Apply applies the rule once to name. In the Regex style only the first match
// of the pattern is replaced and the rest of name is kept, like the Java
// String.replaceFirst, so a name the pattern does not match is returned
// unchanged. In the MessageFormat style every "{0}" of the rule is replaced by
// name. The zero Rule returns name unchanged.
func (r Rule) Apply(name string) string {
	switch r.style {
	case Regex:
		if r.pattern == nil {
			return name
		}
		loc := r.pattern.FindStringSubmatchIndex(name)
		if loc == nil {
			return name
		}
		var result []byte
		result = append(result, name[:loc[0]]...)
		result = r.pattern.Expand(result, []byte(r.template), []byte(name), loc)
		result = append(result, name[loc[1]:]...)
		return string(result)
	case MessageFormat:
		return strings.ReplaceAll(r.template, placeholder, name)
	default:
		return name
	}
}

// Resolve returns name renamed by the rule found with lookup(key), or by
// defaultSpec when lookup does not find the key. An empty spec, found or
// default, returns name unchanged. The lookup is bound by the caller to the
// annotations of the element and of its ancestors. An invalid spec returns
// name unchanged and the parse error.
func Resolve(lookup func(key string) (string, bool), name, key, defaultSpec string, style Style) (string, error) {
	spec, found := "", false
	if lookup != nil {
		spec, found = lookup(key)
	}
	if !found {
		spec = defaultSpec
	}
	if spec == "" {
		return name, nil
	}
	rule, err := ParseRule(spec, style)
	if err != nil {
		return name, fmt.Errorf("key %q: %w", key, err)
	}
	return rule.Apply(name), nil
}

// javaReplacementToTemplate translates a Java replacement string (the second
// argument of String.replaceFirst) into a Go regexp.Expand template. Java reads
// "$1Impl" as group 1 followed by "Impl" while Go would read it as a group
// named "1Impl", so each "$N" becomes "${N}"; Java takes as many digits as
// name an existing group (with numGroups=2, "$12" is group 1 then "2");
// "${name}" is kept; "\x" is the literal x; a literal '$' becomes "$$".
func javaReplacementToTemplate(replacement string, numGroups int) (string, error) {
	var template strings.Builder
	for i := 0; i < len(replacement); i++ {
		switch c := replacement[i]; c {
		case '\\':
			i++
			if i >= len(replacement) {
				return "", errors.New("character to be escaped is missing in the replacement")
			}
			if replacement[i] == '$' {
				template.WriteString("$$")
			} else {
				template.WriteByte(replacement[i])
			}
		case '$':
			i++
			if i >= len(replacement) {
				return "", errors.New("illegal group reference: group index is missing in the replacement")
			}
			switch {
			case replacement[i] == '{':
				end := strings.IndexByte(replacement[i:], '}')
				if end < 0 {
					return "", errors.New("named capturing group is missing trailing '}' in the replacement")
				}
				groupName := replacement[i+1 : i+end]
				if groupName == "" {
					return "", errors.New("named capturing group has 0 length name in the replacement")
				}
				if _, err := groupIndex(groupName, numGroups); err != nil {
					return "", err
				}
				template.WriteString("${" + groupName + "}")
				i += end
			case isDigit(replacement[i]):
				group := int(replacement[i] - '0')
				if group > numGroups {
					return "", fmt.Errorf("no group %d in the pattern", group)
				}
				for i+1 < len(replacement) && isDigit(replacement[i+1]) {
					next := group*10 + int(replacement[i+1]-'0')
					if next > numGroups {
						break
					}
					group = next
					i++
				}
				fmt.Fprintf(&template, "${%d}", group)
			default:
				return "", errors.New("illegal group reference in the replacement")
			}
		default:
			template.WriteByte(c)
		}
	}
	return template.String(), nil
}

// groupIndex validates a "${name}" group reference: a numeric name must be an
// existing group index and a symbolic name is left to regexp.Expand (an
// unknown name expands to the empty string, as Go does).
func groupIndex(name string, numGroups int) (int, error) {
	index := 0
	for _, c := range []byte(name) {
		if !isDigit(c) {
			return -1, nil
		}
		index = index*10 + int(c-'0')
	}
	if index > numGroups {
		return -1, fmt.Errorf("no group %d in the pattern", index)
	}
	return index, nil
}

// isDigit reports whether c is an ASCII digit.
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
