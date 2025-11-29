package commands

import (
	"testing"

	"github.com/iltempo/interplay/sequence"
)

// mockVerboseController implements VerboseController for testing
type mockVerboseController struct {
	verbose bool
}

func (m *mockVerboseController) SetVerbose(v bool) {
	m.verbose = v
}

func (m *mockVerboseController) IsVerbose() bool {
	return m.verbose
}

// TestHandleSet tests the set command
func TestHandleSet(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Valid set command
	err := handler.ProcessCommand("set 1 C4")
	if err != nil {
		t.Errorf("ProcessCommand('set 1 C4') unexpected error: %v", err)
	}
	step, _ := pattern.GetStep(1)
	if step.Note != 60 || step.IsRest {
		t.Errorf("set 1 C4 did not set note correctly")
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("set 1")
	if err == nil {
		t.Error("ProcessCommand('set 1') should return error")
	}

	// Invalid: bad step number
	err = handler.ProcessCommand("set abc C4")
	if err == nil {
		t.Error("ProcessCommand('set abc C4') should return error")
	}

	// Invalid: bad note name
	err = handler.ProcessCommand("set 1 X99")
	if err == nil {
		t.Error("ProcessCommand('set 1 X99') should return error")
	}

	// Invalid: step out of range
	err = handler.ProcessCommand("set 17 C4")
	if err == nil {
		t.Error("ProcessCommand('set 17 C4') should return error")
	}
}

// TestHandleRest tests the rest command
func TestHandleRest(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Set a note first
	pattern.SetNote(1, 60)

	// Valid rest command
	err := handler.ProcessCommand("rest 1")
	if err != nil {
		t.Errorf("ProcessCommand('rest 1') unexpected error: %v", err)
	}
	step, _ := pattern.GetStep(1)
	if !step.IsRest {
		t.Error("rest 1 did not set step to rest")
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("rest")
	if err == nil {
		t.Error("ProcessCommand('rest') should return error")
	}

	// Invalid: bad step number
	err = handler.ProcessCommand("rest abc")
	if err == nil {
		t.Error("ProcessCommand('rest abc') should return error")
	}
}

// TestHandleClear tests the clear command
func TestHandleClear(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Set some notes
	pattern.SetNote(1, 60)
	pattern.SetNote(5, 67)

	// Clear
	err := handler.ProcessCommand("clear")
	if err != nil {
		t.Errorf("ProcessCommand('clear') unexpected error: %v", err)
	}

	// Check all steps are rests
	for i := 1; i <= sequence.NumSteps; i++ {
		step, _ := pattern.GetStep(i)
		if !step.IsRest {
			t.Errorf("After clear, step %d is not a rest", i)
		}
	}
}

// TestHandleReset tests the reset command
func TestHandleReset(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Clear the pattern first
	pattern.Clear()

	// Reset
	err := handler.ProcessCommand("reset")
	if err != nil {
		t.Errorf("ProcessCommand('reset') unexpected error: %v", err)
	}

	// Check that pattern has the default notes
	step1, _ := pattern.GetStep(1)
	if step1.IsRest || step1.Note != 48 {
		t.Error("After reset, step 1 should be C3")
	}
}

// TestHandleVelocity tests the velocity command
func TestHandleVelocity(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Valid velocity
	err := handler.ProcessCommand("velocity 1 80")
	if err != nil {
		t.Errorf("ProcessCommand('velocity 1 80') unexpected error: %v", err)
	}
	step, _ := pattern.GetStep(1)
	if step.Velocity != 80 {
		t.Errorf("velocity 1 80 set velocity to %d, want 80", step.Velocity)
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("velocity 1")
	if err == nil {
		t.Error("ProcessCommand('velocity 1') should return error")
	}

	// Invalid: velocity out of range
	err = handler.ProcessCommand("velocity 1 128")
	if err == nil {
		t.Error("ProcessCommand('velocity 1 128') should return error")
	}

	// Invalid: bad step number
	err = handler.ProcessCommand("velocity abc 80")
	if err == nil {
		t.Error("ProcessCommand('velocity abc 80') should return error")
	}
}

// TestHandleGate tests the gate command
func TestHandleGate(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Valid gate
	err := handler.ProcessCommand("gate 1 50")
	if err != nil {
		t.Errorf("ProcessCommand('gate 1 50') unexpected error: %v", err)
	}
	step, _ := pattern.GetStep(1)
	if step.Gate != 50 {
		t.Errorf("gate 1 50 set gate to %d, want 50", step.Gate)
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("gate 1")
	if err == nil {
		t.Error("ProcessCommand('gate 1') should return error")
	}

	// Invalid: gate out of range (too low)
	err = handler.ProcessCommand("gate 1 0")
	if err == nil {
		t.Error("ProcessCommand('gate 1 0') should return error")
	}

	// Invalid: gate out of range (too high)
	err = handler.ProcessCommand("gate 1 101")
	if err == nil {
		t.Error("ProcessCommand('gate 1 101') should return error")
	}
}

// TestHandleTempo tests the tempo command
func TestHandleTempo(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Valid tempo
	err := handler.ProcessCommand("tempo 120")
	if err != nil {
		t.Errorf("ProcessCommand('tempo 120') unexpected error: %v", err)
	}
	if pattern.GetBPM() != 120 {
		t.Errorf("tempo 120 set BPM to %d, want 120", pattern.GetBPM())
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("tempo")
	if err == nil {
		t.Error("ProcessCommand('tempo') should return error")
	}

	// Invalid: tempo out of range
	err = handler.ProcessCommand("tempo 19")
	if err == nil {
		t.Error("ProcessCommand('tempo 19') should return error")
	}

	// Invalid: bad BPM value
	err = handler.ProcessCommand("tempo abc")
	if err == nil {
		t.Error("ProcessCommand('tempo abc') should return error")
	}
}

// TestUnknownCommand tests handling of unknown commands
func TestUnknownCommand(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	err := handler.ProcessCommand("unknowncommand")
	if err == nil {
		t.Error("ProcessCommand('unknowncommand') should return error")
	}
}

// TestEmptyCommand tests handling of empty input
func TestEmptyCommand(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Empty string should show pattern (no error)
	err := handler.ProcessCommand("")
	if err != nil {
		t.Errorf("ProcessCommand('') unexpected error: %v", err)
	}
}

// TestCommandCaseSensitivity tests that commands are case-insensitive
func TestCommandCaseSensitivity(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Test uppercase
	err := handler.ProcessCommand("SET 1 C4")
	if err != nil {
		t.Errorf("ProcessCommand('SET 1 C4') unexpected error: %v", err)
	}

	// Test mixed case
	err = handler.ProcessCommand("TeMpO 100")
	if err != nil {
		t.Errorf("ProcessCommand('TeMpO 100') unexpected error: %v", err)
	}
	if pattern.GetBPM() != 100 {
		t.Error("Mixed case tempo command did not work")
	}
}

// TestVelocityPreservation tests that velocity is preserved when changing notes
func TestVelocityPreservation(t *testing.T) {
	pattern := sequence.New()
	handler := New(pattern, &mockVerboseController{})

	// Set note with velocity
	handler.ProcessCommand("set 1 C4")
	handler.ProcessCommand("velocity 1 127")

	step, _ := pattern.GetStep(1)
	if step.Velocity != 127 {
		t.Error("Velocity not set correctly")
	}

	// Change note, velocity should be preserved
	handler.ProcessCommand("set 1 D4")

	step, _ = pattern.GetStep(1)
	if step.Velocity != 127 {
		t.Error("Velocity not preserved when changing note")
	}
	if step.Note != 62 { // D4 = 62
		t.Error("Note not changed correctly")
	}
}
