package template

import "context"

type Template interface {
	// Do all preparing actions, e.g. scan file.
	// Should be called first.
	Prepare(ctx context.Context) error
	// Default relative path for template (=file)
	DefaultPath() string
	// Main render function, where template produce code.
	Render(ctx context.Context)
}
