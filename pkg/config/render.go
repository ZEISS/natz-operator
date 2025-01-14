package config

import "io"

var _ io.Writer = (*writer)(nil)

// Opt is a functional option for the writer.
type Opt func(*writer)

type writer struct {
	len int
	tab int
	cfg *Config
}

// WithTabSize sets the tab size for the writer.
func WithTabSize(tab int) Opt {
	return func(w *writer) {
		w.tab = tab
	}
}

// WothLineLen sets the line length for the writer.
func WithLineLen(len int) Opt {
	return func(w *writer) {
		w.len = len
	}
}

// NewWriter returns a new writer.
func NewWriter(cfg *Config, opts ...Opt) *writer {
	w := &writer{
		cfg: cfg,
		tab: DefaulTabSize,
		len: DefaultLineLen,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Write implements the io.Writer interface.
func (w *writer) Write(p []byte) (n int, err error) {
	return len(p), nil
}
