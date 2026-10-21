// Package engine runs a generation: it loads the model, resolves the properties,
// loads the templates of the language, asks each selected entry point for the
// files to generate, renders them and writes them.
//
// A language is a value implementing Language, registered in Register from an
// init function. It contributes only the facts-to-text part of a generation: its
// property declarations, its entry points (the templates of the -t option), the
// files each entry point produces, their view models and its template files.
package engine
