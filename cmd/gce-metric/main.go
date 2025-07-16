// This package provides an executable that can generate synthetic metrics for
// Google Cloud that match a profile.
package main

import (
	"os"

	"github.com/go-logr/logr"
)

// The default logr sink; this will be changed as command options are processed.
var logger = logr.Discard() //nolint:gochecknoglobals // The logger is deliberately global

func main() {
	rootCmd, err := newRootCmd()
	if err != nil {
		logger.Error(err, "Error building commands")
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err, "Error executing command")
	}
}
