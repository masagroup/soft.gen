package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleRegex(t *testing.T) {
	tests := []struct {
		spec     string
		name     string
		expected string
	}{
		// bare replacement: pattern "(.*)"
		{"$1Impl", "Book", "BookImpl"},
		{"${1}Impl", "Book", "BookImpl"},
		{"Mock$1", "Book", "MockBook"},
		{"$1Factory", "library", "libraryFactory"},
		{"$1_impl.go", "Book", "Book_impl.go"},
		{"$1", "Book", "Book"},
		{"$0Impl", "Book", "BookImpl"},
		{"Fixed", "Book", "Fixed"},
		{"$1Impl", "", "Impl"},
		{"\\$1", "Book", "$1"},
		{"a\\\\b$1", "Book", "a\\bBook"},
		// pattern:replacement
		{"(.*)Entry:$1MapEntry", "EStringToStringEntry", "EStringToStringMapEntry"},
		{"(.*)Entry:$1MapEntry", "Book", "Book"},
		{"Book:Novel", "BookShelf", "NovelShelf"},
		{"Book:Novel", "BookBook", "NovelBook"}, // replace-first: the second match is kept
		{"Shelf:Case", "BookShelf", "BookCase"},
		{"^E(.*)$:$1", "EClass", "Class"},
		{"(\\w+)\\.(\\w+):$2.$1", "a.b", "b.a"},
		{"(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)(k):$11", "abcdefghijk", "k"},
		{"(a)(b):$12", "ab", "a2"}, // Java takes the digits as long as the group exists
		{"(?P<base>.*)Impl:${base}", "BookImpl", "Book"},
		{"o:0", "Book", "B0ok"},
		{"x*:y", "abc", "yabc"},         // empty first match at 0
		{"Z:Y", "Book", "Book"},         // no match
		{":Empty", "Book", "EmptyBook"}, // empty pattern: empty first match at 0
		{"Book:", "BookShelf", "Shelf"},
		{"a:b:c", "a", "b:c"}, // split on the first ':' only
	}
	for _, test := range tests {
		t.Run(test.spec+"/"+test.name, func(t *testing.T) {
			rule, err := ParseRule(test.spec, Regex)
			require.NoError(t, err)
			assert.Equal(t, Regex, rule.Style())
			assert.Equal(t, test.expected, rule.Apply(test.name))
		})
	}
}

func TestRuleRegexErrors(t *testing.T) {
	tests := []struct {
		spec string
		msg  string
	}{
		{"(.*:$1", "missing closing )"},
		{"[a-:$1", "missing closing ]"},
		{"$2Impl", "no group 2"},
		{"(.*):$2", "no group 2"},
		{"(.*):${2}", "no group 2"},
		{"$", "group index is missing"},
		{"$Impl", "illegal group reference"},
		{"$$1", "illegal group reference"}, // "$$" is not an escape in Java
		{"${", "missing trailing '}'"},
		{"${}", "0 length name"},
		{"Impl\\", "character to be escaped is missing"},
	}
	for _, test := range tests {
		t.Run(test.spec, func(t *testing.T) {
			_, err := ParseRule(test.spec, Regex)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.spec)
			assert.Contains(t, err.Error(), test.msg)
		})
	}
}

func TestRuleMessageFormat(t *testing.T) {
	tests := []struct {
		spec     string
		name     string
		expected string
	}{
		{"{0}Base", "Book", "BookBase"},
		{"{0}", "Book", "Book"},
		{"Mock{0}", "Book", "MockBook"},
		{"{0}::{0}", "Book", "Book::Book"},
		{"Fixed", "Book", "Fixed"},
		{"$1{0}", "Book", "$1Book"}, // no regex syntax in this style
		{"{0}Base", "", "Base"},
		{"", "Book", ""},
	}
	for _, test := range tests {
		t.Run(test.spec+"/"+test.name, func(t *testing.T) {
			rule, err := ParseRule(test.spec, MessageFormat)
			require.NoError(t, err)
			assert.Equal(t, MessageFormat, rule.Style())
			assert.Equal(t, test.expected, rule.Apply(test.name))
		})
	}
}

func TestRuleMessageFormatErrors(t *testing.T) {
	for _, spec := range []string{"{1}Base", "{0}{1}", "{00}"} {
		t.Run(spec, func(t *testing.T) {
			_, err := ParseRule(spec, MessageFormat)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unknown placeholder")
		})
	}
}

func TestParseRuleUnknownStyle(t *testing.T) {
	_, err := ParseRule("$1Impl", Style(42))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown style Style(42)")
}

func TestStyleString(t *testing.T) {
	assert.Equal(t, "Regex", Regex.String())
	assert.Equal(t, "MessageFormat", MessageFormat.String())
	assert.Equal(t, "Style(7)", Style(7).String())
}

func TestZeroRuleApply(t *testing.T) {
	assert.Equal(t, "Book", Rule{}.Apply("Book"))
}

func TestResolve(t *testing.T) {
	annotations := map[string]string{
		"implementationName": "$1Implementation",
		"mockName":           "(.*):Fake$1",
		"baseName":           "{0}Abstract",
		"emptyName":          "",
		"badName":            "(.*:$1",
	}
	lookup := func(key string) (string, bool) {
		value, found := annotations[key]
		return value, found
	}
	tests := []struct {
		key         string
		defaultSpec string
		style       Style
		expected    string
		err         string
	}{
		// annotation found
		{"implementationName", "$1Impl", Regex, "BookImplementation", ""},
		{"mockName", "Mock$1", Regex, "FakeBook", ""},
		{"baseName", "{0}Base", MessageFormat, "BookAbstract", ""},
		// annotation not found: default
		{"extensionName", "$1Ext", Regex, "BookExt", ""},
		{"extensionName", "{0}Ext", MessageFormat, "BookExt", ""},
		{"interfaceName", "", Regex, "Book", ""},
		{"interfaceName", "", MessageFormat, "Book", ""},
		// empty annotation: name unchanged
		{"emptyName", "$1Impl", Regex, "Book", ""},
		// invalid annotation
		{"badName", "$1Impl", Regex, "Book", `key "badName": naming: rule "(.*:$1": error parsing regexp`},
	}
	for _, test := range tests {
		t.Run(test.key+"/"+test.defaultSpec, func(t *testing.T) {
			result, err := Resolve(lookup, "Book", test.key, test.defaultSpec, test.style)
			assert.Equal(t, test.expected, result)
			if test.err == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
			}
		})
	}
}

func TestResolveNilLookup(t *testing.T) {
	result, err := Resolve(nil, "Book", "implementationName", "$1Impl", Regex)
	require.NoError(t, err)
	assert.Equal(t, "BookImpl", result)

	result, err = Resolve(nil, "Book", "implementationName", "", Regex)
	require.NoError(t, err)
	assert.Equal(t, "Book", result)

	_, err = Resolve(nil, "Book", "implementationName", "$3", Regex)
	require.Error(t, err)
}

func TestResolveLookupAncestors(t *testing.T) {
	// The caller binds lookup to the element and its ancestors: here the
	// package annotation applies when the class has none.
	classAnnotations := map[string]string{}
	packageAnnotations := map[string]string{"implementationName": "$1Object"}
	lookup := func(key string) (string, bool) {
		if value, found := classAnnotations[key]; found {
			return value, true
		}
		value, found := packageAnnotations[key]
		return value, found
	}
	result, err := Resolve(lookup, "Book", "implementationName", "$1Impl", Regex)
	require.NoError(t, err)
	assert.Equal(t, "BookObject", result)

	classAnnotations["implementationName"] = "$1Entity"
	result, err = Resolve(lookup, "Book", "implementationName", "$1Impl", Regex)
	require.NoError(t, err)
	assert.Equal(t, "BookEntity", result)
}
