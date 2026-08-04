// Package logger provides structured logging functionality using Go's log/slog package.
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Options defines configuration options for creating a new slog.Logger instance.
type Options struct {
	// Verbose enables debug level logging when true.
	Verbose bool
	// JSONFormat uses slog.JSONHandler when true; otherwise uses slog.TextHandler.
	JSONFormat bool
	// Output destination for log entries. Defaults to os.Stdout if nil.
	Output io.Writer
}

// New creates and configures a new *slog.Logger instance based on provided options.
func New(opts Options) *slog.Logger {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	level := slog.LevelInfo
	if opts.Verbose {
		level = slog.LevelDebug
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if opts.JSONFormat {
		handler = slog.NewJSONHandler(out, handlerOpts)
	} else {
		handler = slog.NewTextHandler(out, handlerOpts)
	}

	return slog.New(handler)
}
