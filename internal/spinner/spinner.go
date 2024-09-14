package spinner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theckman/yacspin"
)

// CreateSpinner creates and configures a new spinner using the yacspin library.
// The spinner is customized with specific settings like character set, colors, and stop characters for success and failure.
// Parameters: None.
// Returns:
// - A pointer to the initialized yacspin.Spinner.
// - An error if the spinner could not be created.
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

// StopOnSignal sets up a goroutine to stop the spinner and exit the program when an interrupt or termination signal is received.
// It listens for OS signals like SIGINT and SIGTERM, and if such a signal is detected, it stops the spinner with a failure message and exits the program.
// Parameters:
// - spinner: A pointer to the yacspin.Spinner that should be stopped on receiving a signal.
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
