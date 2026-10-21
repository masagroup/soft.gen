package writer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// RegionSyntax is the syntax of a user code region: a start marker line holding
// the region id, and an end marker line. Start is a format with one %s for the
// id; both markers may be preceded by indentation and a comment prefix.
//
// The regions of the Acceleo generators are written as
//
//	// Start of user code <id>
//	...kept on regeneration...
//	// End of user code
type RegionSyntax struct {
	Start string
	End   string
}

// Open returns the start marker of the region id.
func (s RegionSyntax) Open(id string) string { return fmt.Sprintf(s.Start, id) }

// Close returns the end marker.
func (s RegionSyntax) Close() string { return s.End }

// region is one user code region of a text: its id and the lines between its markers.
type region struct {
	id    string
	inner []string
}

// PreserveRegions returns generated with, for every user code region it
// declares, the content the same region has in existing. It also returns the
// ids of the regions of existing that generated does not declare any more
// (their content is lost).
func PreserveRegions(syntax RegionSyntax, existing, generated []byte) ([]byte, []string, error) {
	old, err := parseRegions(syntax, existing)
	if err != nil {
		return nil, nil, fmt.Errorf("existing file: %w", err)
	}
	if len(old) == 0 {
		return generated, nil, nil
	}
	start := startPattern(syntax)
	end := strings.TrimSpace(syntax.End)
	var out []string
	seen := map[string]bool{}
	lines := strings.Split(string(generated), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		out = append(out, line)
		m := start.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id := m[1]
		seen[id] = true
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) != end {
			j++
		}
		if j == len(lines) {
			return nil, nil, fmt.Errorf("generated file: region %q has no end marker", id)
		}
		if kept, ok := old[id]; ok {
			out = append(out, kept.inner...)
		} else {
			out = append(out, lines[i+1:j]...)
		}
		out = append(out, lines[j])
		i = j
	}
	var lost []string
	for id := range old {
		if !seen[id] {
			lost = append(lost, id)
		}
	}
	sortStrings(lost)
	return []byte(strings.Join(out, "\n")), lost, nil
}

// parseRegions indexes the regions of a text by id.
func parseRegions(syntax RegionSyntax, text []byte) (map[string]region, error) {
	regions := map[string]region{}
	if !bytes.Contains(text, []byte(strings.TrimSpace(syntax.End))) {
		return regions, nil
	}
	start := startPattern(syntax)
	end := strings.TrimSpace(syntax.End)
	lines := strings.Split(string(text), "\n")
	for i := 0; i < len(lines); i++ {
		m := start.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		id := m[1]
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) != end {
			j++
		}
		if j == len(lines) {
			return nil, fmt.Errorf("region %q has no end marker", id)
		}
		regions[id] = region{id: id, inner: lines[i+1 : j]}
		i = j
	}
	return regions, nil
}

// startPattern turns the start format into a pattern capturing the id.
func startPattern(syntax RegionSyntax) *regexp.Regexp {
	parts := strings.SplitN(syntax.Start, "%s", 2)
	if len(parts) != 2 {
		return regexp.MustCompile(regexp.QuoteMeta(strings.TrimSpace(syntax.Start)) + `()`)
	}
	return regexp.MustCompile(regexp.QuoteMeta(strings.TrimLeft(parts[0], " \t")) + `(.*?)` + regexp.QuoteMeta(strings.TrimRight(parts[1], " \t")) + `\s*$`)
}
