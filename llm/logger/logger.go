// Package logger provides logging functionality for the browser.
package logger

import (
	"io"
	"log"
	"os"
)

// Logger for browser operations.
// Set to nil to disable logging, or configure with log.SetOutput/SetFlags.
var Logger *log.Logger

func init() {
	// Enable logging by default (for development)
	// Disable only if PRODUCTION environment variable is set
	if os.Getenv("PRODUCTION") != "" {
		Logger = log.New(io.Discard, "", 0) // Silent in production
	} else {
		Logger = log.New(os.Stderr, "[HTTP] ", log.Ltime) // Verbose by default
	}
}
