package commands

import (
	"fmt"
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
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	err = handler.ProcessCommand(fmt.Sprintf("set %d C4", pattern.Length()+1))
	if err == nil {
		t.Errorf("ProcessCommand('set %d C4') should return error", pattern.Length()+1)
	}
}

// TestHandleRest tests the rest command
func TestHandleRest(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	for i := 1; i <= pattern.Length(); i++ {
		step, _ := pattern.GetStep(i)
		if !step.IsRest {
			t.Errorf("After clear, step %d is not a rest", i)
		}
	}
}

// TestHandleReset tests the reset command
func TestHandleReset(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
	handler := New(pattern, &mockVerboseController{})

	// Clear the pattern first
	pattern.Clear()

	// Reset
	err := handler.ProcessCommand("reset")
	if err != nil {
		t.Errorf("ProcessCommand('reset') unexpected error: %v", err)
	}

	// Check that pattern is reset to default (all rests, clean slate)
	step1, _ := pattern.GetStep(1)
	if !step1.IsRest {
		t.Errorf("After reset, step 1 should be a rest, got note %d", step1.Note)
	}
	// Also verify the pattern length is reset to default
	if pattern.Length() != sequence.DefaultPatternLength {
		t.Errorf("After reset, pattern length should be %d, got %d", sequence.DefaultPatternLength, pattern.Length())
	}
}

// TestHandleVelocity tests the velocity command
func TestHandleVelocity(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	pattern := sequence.New(sequence.DefaultPatternLength)
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

// TestHandleLength tests the length command
func TestHandleLength(t *testing.T) {
	initialLength := 4
	pattern := sequence.New(initialLength)
	handler := New(pattern, &mockVerboseController{})

	// Set a note in the initial pattern
	pattern.SetNote(1, 60) // C4
	step1, _ := pattern.GetStep(1)
	if step1.Note != 60 {
		t.Errorf("Initial note not set correctly, got %d", step1.Note)
	}

	// Test valid length change (expansion)
	newLength := 8
	err := handler.ProcessCommand(fmt.Sprintf("length %d", newLength))
	if err != nil {
		t.Errorf("ProcessCommand('length %d') unexpected error: %v", newLength, err)
	}
	if pattern.Length() != newLength {
		t.Errorf("Pattern length after resize = %d, want %d", pattern.Length(), newLength)
	}
	// Check if existing note is preserved
	step1AfterResize, _ := pattern.GetStep(1)
	if step1AfterResize.Note != 60 {
		t.Errorf("Existing note not preserved after expansion, got %d", step1AfterResize.Note)
	}
	// Check if new steps are rests
	for i := initialLength; i < newLength; i++ {
		step, _ := pattern.GetStep(i + 1) // 1-indexed
		if !step.IsRest {
			t.Errorf("New step %d is not a rest after expansion", i+1)
		}
	}

	// Test valid length change (truncation)
	newLength = 2
	err = handler.ProcessCommand(fmt.Sprintf("length %d", newLength))
	if err != nil {
		t.Errorf("ProcessCommand('length %d') unexpected error: %v", newLength, err)
	}
	if pattern.Length() != newLength {
		t.Errorf("Pattern length after truncation = %d, want %d", pattern.Length(), newLength)
	}
	// Check if existing note is still preserved
	step1AfterTruncation, _ := pattern.GetStep(1)
	if step1AfterTruncation.Note != 60 {
		t.Errorf("Existing note not preserved after truncation, got %d", step1AfterTruncation.Note)
	}

	// Invalid: too few arguments
	err = handler.ProcessCommand("length")
	if err == nil {
		t.Error("ProcessCommand('length') with no arguments should return error")
	}

	// Invalid: non-numeric length
	err = handler.ProcessCommand("length abc")
	if err == nil {
		t.Error("ProcessCommand('length abc') should return error")
	}

	// Invalid: zero length
	err = handler.ProcessCommand("length 0")
	if err == nil {
		t.Error("ProcessCommand('length 0') should return error")
	}

	// Invalid: negative length
	err = handler.ProcessCommand("length -5")
	if err == nil {
		t.Error("ProcessCommand('length -5') should return error")
	}
}

// TestHandleTempo tests the tempo command
func TestHandleTempo(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	pattern := sequence.New(sequence.DefaultPatternLength)
	handler := New(pattern, &mockVerboseController{})

	err := handler.ProcessCommand("unknowncommand")
	if err == nil {
		t.Error("ProcessCommand('unknowncommand') should return error")
	}
}

// TestEmptyCommand tests handling of empty input
func TestEmptyCommand(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
	handler := New(pattern, &mockVerboseController{})

	// Empty string should show pattern (no error)
	err := handler.ProcessCommand("")
	if err != nil {
		t.Errorf("ProcessCommand('') unexpected error: %v", err)
	}
}

// TestCommandCaseSensitivity tests that commands are case-insensitive
func TestCommandCaseSensitivity(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
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
	pattern := sequence.New(sequence.DefaultPatternLength)
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

// TestSetRest tests setting a step to rest using the set command
func TestSetRest(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
	handler := New(pattern, &mockVerboseController{})

	// Set a note first
	pattern.SetNote(1, 60)
	step, _ := pattern.GetStep(1)
	if step.IsRest {
		t.Error("Step should have a note initially")
	}

	// Use set command to set rest
	err := handler.ProcessCommand("set 1 rest")
	if err != nil {
		t.Errorf("ProcessCommand('set 1 rest') unexpected error: %v", err)
	}

	step, _ = pattern.GetStep(1)
	if !step.IsRest {
		t.Error("set 1 rest should set step to rest")
	}

	// Test case insensitive
	pattern.SetNote(2, 60)
	err = handler.ProcessCommand("set 2 REST")
	if err != nil {
		t.Errorf("ProcessCommand('set 2 REST') unexpected error: %v", err)
	}

	step, _ = pattern.GetStep(2)
	if !step.IsRest {
		t.Error("set 2 REST should set step to rest (case insensitive)")
	}
}

// TestSetWithParameters tests the set command with vel, gate, and dur parameters
func TestSetWithParameters(t *testing.T) {
	pattern := sequence.New(sequence.DefaultPatternLength)
	handler := New(pattern, &mockVerboseController{})

	// Test set with all parameters
	err := handler.ProcessCommand("set 1 C4 vel:120 gate:85 dur:3")
	if err != nil {
		t.Errorf("ProcessCommand('set 1 C4 vel:120 gate:85 dur:3') unexpected error: %v", err)
	}

	step, _ := pattern.GetStep(1)
	if step.Note != 60 {
		t.Errorf("Note not set correctly, got %d, want 60", step.Note)
	}
	if step.Velocity != 120 {
		t.Errorf("Velocity not set correctly, got %d, want 120", step.Velocity)
	}
	if step.Gate != 85 {
		t.Errorf("Gate not set correctly, got %d, want 85", step.Gate)
	}
	if step.Duration != 3 {
		t.Errorf("Duration not set correctly, got %d, want 3", step.Duration)
	}

	// Test set with only vel parameter
	err = handler.ProcessCommand("set 2 D4 vel:100")
	if err != nil {
		t.Errorf("ProcessCommand('set 2 D4 vel:100') unexpected error: %v", err)
	}
	step, _ = pattern.GetStep(2)
	if step.Note != 62 || step.Velocity != 100 {
		t.Error("Set with vel parameter failed")
	}

	// Test set with only gate parameter
	err = handler.ProcessCommand("set 3 E4 gate:50")
	if err != nil {
		t.Errorf("ProcessCommand('set 3 E4 gate:50') unexpected error: %v", err)
	}
	step, _ = pattern.GetStep(3)
	if step.Note != 64 || step.Gate != 50 {
		t.Error("Set with gate parameter failed")
	}

	// Test set with only dur parameter
	err = handler.ProcessCommand("set 4 F4 dur:2")
	if err != nil {
		t.Errorf("ProcessCommand('set 4 F4 dur:2') unexpected error: %v", err)
	}
	step, _ = pattern.GetStep(4)
	if step.Note != 65 || step.Duration != 2 {
		t.Error("Set with dur parameter failed")
	}

	// Test parameters in different order
	err = handler.ProcessCommand("set 5 G4 dur:4 vel:110 gate:75")
	if err != nil {
		t.Errorf("ProcessCommand('set 5 G4 dur:4 vel:110 gate:75') unexpected error: %v", err)
	}
	step, _ = pattern.GetStep(5)
	if step.Note != 67 || step.Velocity != 110 || step.Gate != 75 || step.Duration != 4 {
		t.Error("Set with parameters in different order failed")
	}

	// Test invalid velocity
	err = handler.ProcessCommand("set 6 A4 vel:128")
	if err == nil {
		t.Error("ProcessCommand('set 6 A4 vel:128') should return error (velocity too high)")
	}

	// Test invalid gate
	err = handler.ProcessCommand("set 6 A4 gate:101")
	if err == nil {
		t.Error("ProcessCommand('set 6 A4 gate:101') should return error (gate too high)")
	}

	// Test invalid parameter name
	err = handler.ProcessCommand("set 6 A4 invalid:50")
	if err == nil {
		t.Error("ProcessCommand('set 6 A4 invalid:50') should return error (unknown parameter)")
	}
}
