package tmpl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/masagroup/soft.gen/v2/naming"
)

// BaseFuncs returns the presentational functions every template set has.
// None of them knows the model: a decision about the model is taken in Go, in
// the view model, never in a template.
//
//	include name data   execute the named template, return its text (dynamic dispatch)
//	blank               the empty string: a whitespace barrier that keeps a blank line before a {{- if }} line
//	indent n s          prefix every non-empty line of s with n tabs
//	join sep list       join strings with a separator
//	upperFirst s        first letter upper-cased
//	lowerFirst s        first letter lower-cased
//	upperSnake s        EMF constant form (eAllStructuralFeatures -> EALL_STRUCTURAL_FEATURES)
//	lower s / upper s   whole string case
//	quote s             s as a double-quoted literal with escapes
//	list a b c          a slice of the arguments
//	dict k v k v        a map from the pairs of arguments
//	ternary cond a b    a when cond is true, else b
//	default d v         v when it is non-empty, else d
//	add a b / sub a b   integer arithmetic
func BaseFuncs() template.FuncMap {
	return template.FuncMap{
		"blank":      func() string { return "" },
		"indent":     indent,
		"join":       join,
		"upperFirst": naming.UpperFirst,
		"lowerFirst": naming.LowerFirst,
		"upperSnake": naming.UpperSnake,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"quote":      strconv.Quote,
		"list":       func(items ...any) []any { return items },
		"dict":       dict,
		"ternary":    ternary,
		"default":    defaultValue,
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
	}
}

// join joins the items of a slice ([]string or []any) with sep.
func join(sep string, items any) (string, error) {
	rv := reflect.ValueOf(items)
	if rv.Kind() != reflect.Slice {
		return "", fmt.Errorf("join: %T is not a slice", items)
	}
	parts := make([]string, rv.Len())
	for i := range parts {
		parts[i] = fmt.Sprint(rv.Index(i).Interface())
	}
	return strings.Join(parts, sep), nil
}

// indent prefixes every non-empty line of s with n tabs.
func indent(n int, s string) string {
	prefix := strings.Repeat("\t", n)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// dict builds a map from alternating keys and values.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict: odd number of arguments")
	}
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key %v is not a string", pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

// ternary returns a when cond is true, else b.
func ternary(cond bool, a, b any) any {
	if cond {
		return a
	}
	return b
}

// defaultValue returns v unless it is empty (zero value or empty slice), then d.
func defaultValue(d, v any) any {
	if v == nil {
		return d
	}
	rv := reflect.ValueOf(v)
	if rv.IsZero() || (rv.Kind() == reflect.Slice && rv.Len() == 0) {
		return d
	}
	return v
}
