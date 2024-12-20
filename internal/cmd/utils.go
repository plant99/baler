package cmd

import (
	"os"

	"github.com/plant99/baler/internal/baler"
	"github.com/spf13/cobra"
)

type cobraLogger struct {
	cmd     *cobra.Command
	verbose bool
}

func newCobraLogger(cmd *cobra.Command, verbose bool) *cobraLogger {
	return &cobraLogger{
		cmd:     cmd,
		verbose: verbose,
	}
}

func (l *cobraLogger) Info(msg string) {
	l.cmd.Printf("Info: %s\n", msg)
}

func (l *cobraLogger) Warn(msg string) {
	l.cmd.Printf("Warning: %s\n", msg)
}

func (l *cobraLogger) Error(msg string) {
	l.cmd.PrintErrf("Error: %s\n", msg)
}

// TODO: the following function should use cobraLogger
func handleError(cmd *cobra.Command, err error) {
	if err == nil {
		return
	}

	// check if err is BalerError
	if balerErr, ok := baler.IsBalerError(err); ok {
		switch balerErr.Type {
		case baler.ErrorTypeValidation:
			cmd.PrintErrf("Validation error: %v\n", balerErr)
			os.Exit(1)
		case baler.ErrorTypeIO:
			cmd.PrintErrf("I/O error: %v\n", balerErr)
			os.Exit(1)
		case baler.ErrorTypeConfig:
			cmd.PrintErrf("Configuration error: %v\n", balerErr)
			os.Exit(1)
		case baler.ErrorTypeInternal:
			cmd.PrintErrf("Internal error: %v\n", balerErr)
			os.Exit(1)
		}
	}

	cmd.PrintErrf("Unexpected error: %v\n", err)
	os.Exit(1)
}
