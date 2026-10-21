// Package writer writes the generated files: it normalizes the line endings,
// prepends the generated-code header, preserves the user code regions of the
// previous version of a file, formats the content with the formatter of the
// language, creates the directories and logs the written path. It also
// implements the dry-run and diff modes.
package writer
