package spinner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theckman/yacspin"
)

// MockSpinner is a mock implementation of SpinnerInterface
type MockSpinner struct {
	mock.Mock
}

func (m *MockSpinner) Status() yacspin.SpinnerStatus {
	args := m.Called()
	return args.Get(0).(yacspin.SpinnerStatus)
}

func (m *MockSpinner) Message(message string) {
	m.Called(message)
}

func (m *MockSpinner) StopMessage(message string) {
	m.Called(message)
}

func (m *MockSpinner) StopFailMessage(message string) {
	m.Called(message)
}

func (m *MockSpinner) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSpinner) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSpinner) StopFail() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSpinner) New(cfg yacspin.Config) (*yacspin.Spinner, error) {
	args := m.Called(cfg)
	return args.Get(0).(*yacspin.Spinner), args.Error(1)
}

func TestCreateSpinner(t *testing.T) {
	// Save the original defaultSpinnerCreator
	originalCreator := DefaultSpinnerCreator

	// Restore the original defaultSpinnerCreator after the test
	defer func() {
		DefaultSpinnerCreator = originalCreator
	}()

	testCases := []struct {
		name          string
		creatorFunc   func(yacspin.Config) (*yacspin.Spinner, error)
		expectedError bool
	}{
		{
			name: "Successful spinner creation",
			creatorFunc: func(cfg yacspin.Config) (*yacspin.Spinner, error) {
				return &yacspin.Spinner{}, nil
			},
			expectedError: false,
		},
		{
			name: "Failed spinner creation",
			creatorFunc: func(cfg yacspin.Config) (*yacspin.Spinner, error) {
				return nil, assert.AnError
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the defaultSpinnerCreator to our test function
			DefaultSpinnerCreator = tc.creatorFunc

			// Call CreateSpinner
			spinner, err := CreateSpinner()

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, spinner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, spinner)
				assert.IsType(t, &YacspinWrapper{}, spinner)
			}
		})
	}
}

func TestStopOnSignal(t *testing.T) {
	mockSpinner := new(MockSpinner)
	mockSpinner.On("Status").Return(yacspin.SpinnerRunning)
	mockSpinner.On("StopFailMessage", "Interrupted by user").Return()
	mockSpinner.On("StopFail").Return(nil)

	StopOnSignal(mockSpinner)

	// Simulate context cancellation
	cancelFunc()

	// Wait for goroutine to complete
	time.Sleep(300 * time.Millisecond)

	mockSpinner.AssertExpectations(t)
}

func TestStopOnSignalNotRunning(t *testing.T) {
	mockSpinner := new(MockSpinner)
	mockSpinner.On("Status").Return(yacspin.SpinnerStopped)
	mockSpinner.On("StopMessage", "Program stopped").Return()

	StopOnSignal(mockSpinner)

	// Simulate context cancellation
	cancelFunc()

	// Wait for goroutine to complete
	time.Sleep(300 * time.Millisecond)

	mockSpinner.AssertExpectations(t)
}

func TestStopOnSignalError(t *testing.T) {
	mockSpinner := new(MockSpinner)
	mockSpinner.On("Status").Return(yacspin.SpinnerRunning)
	mockSpinner.On("StopFailMessage", "Interrupted by user").Return()
	mockSpinner.On("StopFail").Return(errors.New("stop fail error"))

	StopOnSignal(mockSpinner)

	// Simulate context cancellation
	cancelFunc()

	// Wait for goroutine to complete
	time.Sleep(300 * time.Millisecond)

	mockSpinner.AssertExpectations(t)
}

func TestHandleSignals(t *testing.T) {
	// Save original exitFunc and restore it after the test
	originalExitFunc := exitFunc
	defer func() { exitFunc = originalExitFunc }()

	// Save original sigCh and restore it after the test
	originalSigCh := sigCh
	defer func() {
		sigCh = originalSigCh
		signal.Stop(sigCh)
	}()

	// Create a new sigCh for this test
	sigCh = make(chan os.Signal, 1)

	// Create a channel to track if exitFunc was called
	exitCalled := make(chan int, 1)

	// Mock exitFunc
	exitFunc = func(code int) {
		exitCalled <- code
	}

	// Create a context and cancelFunc
	_, cancel := context.WithCancel(context.Background())

	// Create a channel to track if cancelFunc was called
	cancelCalled := make(chan struct{})
	wrappedCancel := func() {
		cancel()
		close(cancelCalled)
	}

	// Run handleSignals
	handleSignals(wrappedCancel)

	// Allow time for goroutine to start
	time.Sleep(50 * time.Millisecond)

	// Send a signal
	sigCh <- os.Interrupt

	// Check if cancelFunc was called
	select {
	case <-cancelCalled:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("cancelFunc was not called")
	}

	// Check if exitFunc was called
	select {
	case code := <-exitCalled:
		if code != 0 {
			t.Errorf("Expected exit code 0, got %d", code)
		}
	case <-time.After(1 * time.Second):
		t.Error("exitFunc was not called")
	}

	// Cancel the context to clean up
	cancel()
}
