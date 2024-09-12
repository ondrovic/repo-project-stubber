package spinner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theckman/yacspin"
)

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
