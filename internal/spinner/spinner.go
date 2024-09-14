package spinner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theckman/yacspin"
)

// CreateSpinner creates and configures a new yacspin spinner with customized settings.
// The spinner's appearance and behavior are defined in the configuration, including
// the character set, colors, and symbols for both success and failure states.
// It returns the configured spinner and an error if spinner creation fails.
func CreateSpinner() (*yacspin.Spinner, error) {
	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[14],
		Suffix:            " ",
		SuffixAutoColon:   true,
		Colors:            []string{"fgBlue"},
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
	}

	s, err := yacspin.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to make spinner from struct: %w", err)
	}

	return s, nil
}

// StopOnSignal listens for OS interrupt signals (e.g., Ctrl+C) and stops the spinner with
// a failure message if the signal is received. It checks if the spinner is currently running
// and stops it using a fail message. This function ensures proper cleanup when the program
// is interrupted by the user
func StopOnSignal(spinner *yacspin.Spinner) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh

		// Check if the spinner is running by using the Status() method
		if spinner.Status() == yacspin.SpinnerRunning {
			spinner.StopFailMessage("interrupted by user")

			// handle the error if needed
			if err := spinner.StopFail(); err != nil {
				// log or handle the error if stopping fails
				fmt.Println("Error stopping spinner:", err)
			}
		}

		os.Exit(0)
	}()
}
