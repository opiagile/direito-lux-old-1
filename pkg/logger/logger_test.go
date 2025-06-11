package logger

import (
	"testing"
)

func TestInit(t *testing.T) {
	// Test logger initialization with different configurations
	testCases := []struct {
		level    string
		encoding string
		output   string
		wantErr  bool
	}{
		{"info", "json", "stdout", false},
		{"debug", "console", "stdout", false},
		{"error", "json", "stdout", false},
		{"invalid", "json", "stdout", true},
		{"info", "invalid", "stdout", true},
	}
	
	for _, tc := range testCases {
		err := Init(tc.level, tc.encoding, tc.output)
		if (err != nil) != tc.wantErr {
			t.Errorf("Init(%s, %s, %s) error = %v, wantErr %v", 
				tc.level, tc.encoding, tc.output, err, tc.wantErr)
		}
	}
}

func TestLoggerFunctions(t *testing.T) {
	// Initialize logger for testing
	err := Init("info", "json", "stdout")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	
	// Test that logger functions don't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logger function panicked: %v", r)
		}
	}()
	
	Info("test info message")
	Debug("test debug message")
	Error("test error message")
	Warn("test warn message")
}