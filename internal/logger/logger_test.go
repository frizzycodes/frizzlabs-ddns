package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/frizzlabs/frizzlabs-ddns/internal/logger"
)

func TestLoggerTextHandler(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Verbose:    false,
		JSONFormat: false,
		Output:     &buf,
	})

	l.Info("test info message", "key", "value")
	l.Debug("test debug message")

	output := buf.String()
	if !strings.Contains(output, "test info message") {
		t.Errorf("expected output to contain info message, got %q", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected output to contain key=value, got %q", output)
	}
	if strings.Contains(output, "test debug message") {
		t.Errorf("did not expect debug message in non-verbose mode, got %q", output)
	}
}

func TestLoggerVerboseAndJSON(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.Options{
		Verbose:    true,
		JSONFormat: true,
		Output:     &buf,
	})

	l.Debug("debug json message", "component", "test")

	output := buf.String()
	if !strings.Contains(output, `"level":"DEBUG"`) && !strings.Contains(output, `"level":"debug"`) {
		t.Errorf("expected DEBUG level in JSON output, got %q", output)
	}
	if !strings.Contains(output, `"msg":"debug json message"`) {
		t.Errorf("expected message in JSON output, got %q", output)
	}
	if !strings.Contains(output, `"component":"test"`) {
		t.Errorf("expected attribute component=test in JSON output, got %q", output)
	}
}
