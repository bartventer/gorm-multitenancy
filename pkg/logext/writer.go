// Package logext provides a custom logger that logs messages to the provided output.
//
// Logging is controlled by the GMT_DEBUG environment variable; set to
// "true" to enable debug logging.
package logext

import (
	"io"
	"log"
	"os"
)

// DebugEnvVar is the name of the environment variable that controls debug logging.
const DebugEnvVar = "GMT_DEBUG"

// Logger is a custom logger that logs messages to the provided output.
type Logger struct {
	*log.Logger
}

// New returns a new logger.
// If the GMT_DEBUG environment variable is set, it logs messages to the provided output.
// Otherwise, it discards all log messages.
func New(output io.Writer) *Logger {
	if os.Getenv(DebugEnvVar) != "true" {
		output = io.Discard
	}
	return &Logger{log.New(output, "[gorm-multitenancy] ", log.LstdFlags)}
}

var std = New(os.Stderr)

// Default returns the default logger.
func Default() *Logger { return std }
