package writer

import (
	"fmt"
	"sort"
	"strings"
)

// UnifiedDiff returns a unified diff (3 lines of context) between the old and
// new content of the file at path; an empty string when they are equal.
func UnifiedDiff(path string, old, new []byte) string {
	a := splitLines(string(old))
	b := splitLines(string(new))
	ops := diffLines(a, b)
	if len(ops) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n+++ %s\n", path, path)
	const context = 3
	for _, hunk := range hunks(ops, context, len(a), len(b)) {
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", hunk.aStart+1, hunk.aLen, hunk.bStart+1, hunk.bLen)
		for _, op := range hunk.ops {
			switch op.kind {
			case opEqual:
				writeDiffLine(&sb, ' ', a[op.aIndex])
			case opDelete:
				writeDiffLine(&sb, '-', a[op.aIndex])
			case opInsert:
				writeDiffLine(&sb, '+', b[op.bIndex])
			}
		}
	}
	return sb.String()
}

// writeDiffLine writes one line of a hunk: its marker (' ', '-' or '+'), the
// text and a newline.
func writeDiffLine(sb *strings.Builder, marker byte, text string) {
	sb.WriteByte(marker)
	sb.WriteString(text)
	sb.WriteByte('\n')
}

// splitLines splits text into lines without their newline.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.TrimSuffix(text, "\n")
	return strings.Split(text, "\n")
}

type opKind int

const (
	opEqual opKind = iota
	opDelete
	opInsert
)

// op is one line of an edit script.
type op struct {
	kind   opKind
	aIndex int
	bIndex int
}

// diffLines computes an edit script from a to b (longest common subsequence);
// it returns nil when a equals b.
func diffLines(a, b []string) []op {
	n, m := len(a), len(b)
	// lcs[i][j] = length of the LCS of a[i:] and b[j:]
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var ops []op
	changed := false
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, op{opEqual, i, j})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, op{opDelete, i, j})
			changed = true
			i++
		default:
			ops = append(ops, op{opInsert, i, j})
			changed = true
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, op{opDelete, i, j})
		changed = true
	}
	for ; j < m; j++ {
		ops = append(ops, op{opInsert, i, j})
		changed = true
	}
	if !changed {
		return nil
	}
	return ops
}

// hunk is a group of changes with their context lines.
type hunk struct {
	aStart, aLen int
	bStart, bLen int
	ops          []op
}

// hunks groups the edit script into hunks separated by more than 2*context equal lines.
func hunks(ops []op, context, aLen, bLen int) []hunk {
	var result []hunk
	var current *hunk
	equalRun := 0
	flush := func() {
		if current == nil {
			return
		}
		// drop the trailing context beyond `context` lines
		trim := equalRun - context
		if trim > 0 {
			current.ops = current.ops[:len(current.ops)-trim]
		}
		current.aLen, current.bLen = 0, 0
		for _, op := range current.ops {
			if op.kind != opInsert {
				current.aLen++
			}
			if op.kind != opDelete {
				current.bLen++
			}
		}
		result = append(result, *current)
		current = nil
	}
	for idx, op := range ops {
		if op.kind == opEqual {
			if current != nil {
				equalRun++
				current.ops = append(current.ops, op)
				if equalRun > 2*context {
					flush()
				}
			}
			continue
		}
		if current == nil {
			start := idx - context
			if start < 0 {
				start = 0
			}
			current = &hunk{aStart: ops[start].aIndex, bStart: ops[start].bIndex}
			current.ops = append(current.ops, ops[start:idx]...)
		}
		current.ops = append(current.ops, op)
		equalRun = 0
	}
	flush()
	return result
}

// sortStrings sorts in place.
func sortStrings(list []string) { sort.Strings(list) }
