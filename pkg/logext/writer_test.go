package logext

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("debug", func(t *testing.T) {
		var b bytes.Buffer
		t.Setenv(DebugEnvVar, "true")
		logger := New(&b)
		logger.Println("test message")
		if b.String() == "" {
			t.Error("Expected logger to write output")
		}
	})

	t.Run("no debug", func(t *testing.T) {
		var b bytes.Buffer
		logger := New(&b)
		logger.Println("test message")
		if b.String() != "" {
			t.Error("Expected logger to not write output")
		}
	})
}
