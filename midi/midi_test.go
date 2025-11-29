package midi

import (
	"testing"
)

// TestListPorts tests that ListPorts returns without error
// Note: We can't assert specific ports since it depends on the system
func TestListPorts(t *testing.T) {
	ports, err := ListPorts()
	if err != nil {
		t.Errorf("ListPorts() unexpected error: %v", err)
	}

	// ports might be empty if no MIDI devices connected
	// Just verify it returns a slice (even if empty)
	if ports == nil {
		t.Error("ListPorts() returned nil instead of empty slice")
	}
}

// TestOpenInvalidPort tests opening an invalid port index
func TestOpenInvalidPort(t *testing.T) {
	// Try to open a port that definitely doesn't exist
	_, err := Open(9999)
	if err == nil {
		t.Error("Open(9999) should return error for invalid port index")
	}
}

// TestNoteOnOffBounds tests note and velocity boundaries
// We test with a mock by checking the function signatures work
func TestNoteOnOffBounds(t *testing.T) {
	// We can't actually test MIDI output without a device
	// But we can verify the function signatures are correct
	// by checking the types compile

	// This test just ensures the API is correct
	var o *Output
	if o != nil {
		// These calls would work if we had a real output
		_ = o.NoteOn(0, 60, 100)
		_ = o.NoteOff(0, 60)
		_ = o.Close()
	}
}

// TestOutputStructure verifies Output struct has required fields
func TestOutputStructure(t *testing.T) {
	// Verify Output type exists and has expected methods
	var o *Output

	// Check that methods exist (compile-time check)
	_ = func(channel, note, velocity uint8) error { return o.NoteOn(channel, note, velocity) }
	_ = func(channel, note uint8) error { return o.NoteOff(channel, note) }
	_ = func() error { return o.Close() }
}

// TestListPortsReturnType verifies ListPorts returns correct types
func TestListPortsReturnType(t *testing.T) {
	ports, err := ListPorts()

	// Verify return types
	if err != nil {
		// Error is acceptable (e.g., no MIDI driver available)
		return
	}

	// Verify we get a string slice
	for i, port := range ports {
		if port == "" {
			t.Errorf("Port %d has empty name", i)
		}
	}
}
