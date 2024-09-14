package spinner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theckman/yacspin"
)

var (
	globalCtx             context.Context
	cancelFunc            context.CancelFunc
	exitFunc              func(int)      = os.Exit
	sigCh                 chan os.Signal = make(chan os.Signal, 1)
	defaultSpinnerCreator                = yacspin.New
)

// SpinnerInterface defines the methods used from yacspin.Spinner
type SpinnerInterface interface {
	Status() yacspin.SpinnerStatus
	Message(message string)
	StopMessage(message string)
	StopFailMessage(message string)
	Start() error
	Stop() error
	StopFail() error
}

// YacspinWrapper wraps yacspin.Spinner to implement SpinnerInterface
type YacspinWrapper struct {
	*yacspin.Spinner
}

func CreateSpinner() (SpinnerInterface, error) {
	return CreateSpinnerWithCreator(defaultSpinnerCreator)
}

// CreateSpinner creates and configures a new spinner using the yacspin library.
func CreateSpinnerWithCreator(newSpinner func(yacspin.Config) (*yacspin.Spinner, error)) (SpinnerInterface, error) {
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

	s, err := newSpinner(cfg)
	if err != nil {
		// return nil, fmt.Errorf("failed to make spinner from struct: %w", err)
		return nil, err
	}

	return &YacspinWrapper{Spinner: s}, nil
}

func init() {
	// Set up global context and signal handling
	globalCtx, cancelFunc = context.WithCancel(context.Background())
	handleSignals(cancelFunc)
}

func handleSignals(cancelFunc context.CancelFunc) {
	// sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh

		if sig != nil {
			// Cancel the context to signal spinners to stop
			cancelFunc()

			// Give time for the message to be displayed
			time.Sleep(200 * time.Millisecond)

			if exitFunc != nil {
				exitFunc(0)
			}
		}
	}()
}

// StopOnSignal sets up a spinner to stop when the global context is cancelled
func StopOnSignal(spinner SpinnerInterface) {
	go func() {
		<-globalCtx.Done()

		if spinner.Status() == yacspin.SpinnerRunning {
			spinner.StopFailMessage("Interrupted by user")
			if err := spinner.StopFail(); err != nil {
				fmt.Println("Error stopping spinner:", err)
			}
		} else {
			spinner.StopMessage("Program stopped")
		}
	}()
}
