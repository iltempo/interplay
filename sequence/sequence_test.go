package sequence

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNoteNameToMIDI tests note name to MIDI number conversion
func TestNoteNameToMIDI(t *testing.T) {
	tests := []struct {
		name     string
		noteName string
		want     uint8
		wantErr  bool
	}{
		// Valid notes
		{"C4", "C4", 60, false},
		{"A4", "A4", 69, false},
		{"C0", "C0", 12, false},
		{"C3", "C3", 48, false},
		{"G3", "G3", 55, false},

		// Sharps
		{"C#4", "C#4", 61, false},
		{"D#3", "D#3", 51, false},
		{"F#4", "F#4", 66, false},

		// Flats
		{"Db4", "Db4", 61, false},
		{"Eb3", "Eb3", 51, false},
		{"Bb3", "Bb3", 58, false},

		// Edge cases
		{"C8", "C8", 108, false},

		// Invalid inputs
		{"Empty", "", 0, true},
		{"TooShort", "C", 0, true},
		{"InvalidNote", "X4", 0, true},
		{"InvalidOctave", "C99", 0, true},
		{"TooLong", "C#4extra", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NoteNameToMIDI(tt.noteName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoteNameToMIDI(%q) error = %v, wantErr %v", tt.noteName, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NoteNameToMIDI(%q) = %v, want %v", tt.noteName, got, tt.want)
			}
		})
	}
}

// TestMidiToNoteName tests MIDI number to note name conversion
func TestMidiToNoteName(t *testing.T) {
	tests := []struct {
		name string
		note uint8
		want string
	}{
		{"Middle C", 60, "C4"},
		{"A440", 69, "A4"},
		{"Lowest C", 12, "C0"},
		{"C3", 48, "C3"},
		{"G3", 55, "G3"},
		{"C#4", 61, "C#4"},
		{"D#3", 51, "D#3"},
		{"Highest C", 108, "C8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := midiToNoteName(tt.note)
			if got != tt.want {
				t.Errorf("midiToNoteName(%d) = %v, want %v", tt.note, got, tt.want)
			}
		})
	}
}

// TestSetNote tests setting notes on steps
func TestSetNote(t *testing.T) {
	p := New()

	// Valid set
	err := p.SetNote(1, 60)
	if err != nil {
		t.Errorf("SetNote(1, 60) unexpected error: %v", err)
	}
	if p.Steps[0].Note != 60 || p.Steps[0].IsRest {
		t.Errorf("SetNote(1, 60) did not set note correctly")
	}

	// Step out of range (too low)
	err = p.SetNote(0, 60)
	if err == nil {
		t.Error("SetNote(0, 60) should return error for step 0")
	}

	// Step out of range (too high)
	err = p.SetNote(17, 60)
	if err == nil {
		t.Error("SetNote(17, 60) should return error for step 17")
	}

	// Note out of range
	err = p.SetNote(1, 128)
	if err == nil {
		t.Error("SetNote(1, 128) should return error for note > 127")
	}
}

// TestSetVelocity tests velocity setting
func TestSetVelocity(t *testing.T) {
	p := New()

	// Valid velocity
	err := p.SetVelocity(1, 80)
	if err != nil {
		t.Errorf("SetVelocity(1, 80) unexpected error: %v", err)
	}
	if p.Steps[0].Velocity != 80 {
		t.Errorf("SetVelocity(1, 80) got velocity %d, want 80", p.Steps[0].Velocity)
	}

	// Min velocity
	err = p.SetVelocity(1, 0)
	if err != nil {
		t.Errorf("SetVelocity(1, 0) unexpected error: %v", err)
	}

	// Max velocity
	err = p.SetVelocity(1, 127)
	if err != nil {
		t.Errorf("SetVelocity(1, 127) unexpected error: %v", err)
	}

	// Out of range
	err = p.SetVelocity(1, 128)
	if err == nil {
		t.Error("SetVelocity(1, 128) should return error")
	}

	// Invalid step
	err = p.SetVelocity(0, 80)
	if err == nil {
		t.Error("SetVelocity(0, 80) should return error")
	}
}

// TestSetGate tests gate length setting
func TestSetGate(t *testing.T) {
	p := New()

	// Valid gate
	err := p.SetGate(1, 50)
	if err != nil {
		t.Errorf("SetGate(1, 50) unexpected error: %v", err)
	}
	if p.Steps[0].Gate != 50 {
		t.Errorf("SetGate(1, 50) got gate %d, want 50", p.Steps[0].Gate)
	}

	// Min gate
	err = p.SetGate(1, 1)
	if err != nil {
		t.Errorf("SetGate(1, 1) unexpected error: %v", err)
	}

	// Max gate
	err = p.SetGate(1, 100)
	if err != nil {
		t.Errorf("SetGate(1, 100) unexpected error: %v", err)
	}

	// Out of range (too low)
	err = p.SetGate(1, 0)
	if err == nil {
		t.Error("SetGate(1, 0) should return error")
	}

	// Out of range (too high)
	err = p.SetGate(1, 101)
	if err == nil {
		t.Error("SetGate(1, 101) should return error")
	}

	// Invalid step
	err = p.SetGate(0, 50)
	if err == nil {
		t.Error("SetGate(0, 50) should return error")
	}
}

// TestSetRest tests setting rests
func TestSetRest(t *testing.T) {
	p := New()

	// First set a note
	p.SetNote(1, 60)

	// Then set it to rest
	err := p.SetRest(1)
	if err != nil {
		t.Errorf("SetRest(1) unexpected error: %v", err)
	}
	if !p.Steps[0].IsRest {
		t.Error("SetRest(1) did not set IsRest to true")
	}

	// Invalid step
	err = p.SetRest(0)
	if err == nil {
		t.Error("SetRest(0) should return error")
	}
}

// TestClear tests clearing all steps
func TestClear(t *testing.T) {
	p := New()

	// Modify some steps
	p.SetNote(1, 60)
	p.SetNote(5, 67)

	// Clear
	p.Clear()

	// Check all steps are rests
	for i := 0; i < NumSteps; i++ {
		if !p.Steps[i].IsRest {
			t.Errorf("After Clear(), step %d is not a rest", i+1)
		}
	}
}

// TestClone tests pattern cloning
func TestClone(t *testing.T) {
	p := New()
	p.SetNote(1, 60)
	p.SetVelocity(1, 80)
	p.SetGate(1, 50)
	p.SetTempo(120)

	clone := p.Clone()

	// Check values are copied
	if clone.Steps[0].Note != 60 {
		t.Error("Clone did not copy note")
	}
	if clone.Steps[0].Velocity != 80 {
		t.Error("Clone did not copy velocity")
	}
	if clone.Steps[0].Gate != 50 {
		t.Error("Clone did not copy gate")
	}
	if clone.BPM != 120 {
		t.Error("Clone did not copy BPM")
	}

	// Modify clone and ensure original is unchanged
	clone.SetNote(1, 72)
	if p.Steps[0].Note == 72 {
		t.Error("Modifying clone affected original")
	}
}

// TestCopyFrom tests copying from another pattern
func TestCopyFrom(t *testing.T) {
	p1 := New()
	p1.SetNote(1, 60)
	p1.SetVelocity(1, 80)
	p1.SetTempo(120)

	p2 := New()
	p2.CopyFrom(p1)

	// Check values are copied
	if p2.Steps[0].Note != 60 {
		t.Error("CopyFrom did not copy note")
	}
	if p2.Steps[0].Velocity != 80 {
		t.Error("CopyFrom did not copy velocity")
	}
	if p2.BPM != 120 {
		t.Error("CopyFrom did not copy BPM")
	}

	// Modify p1 and ensure p2 is unchanged
	p1.SetNote(1, 72)
	if p2.Steps[0].Note == 72 {
		t.Error("Modifying source after CopyFrom affected destination")
	}
}

// TestSetTempo tests tempo setting
func TestSetTempo(t *testing.T) {
	p := New()

	// Valid tempo
	err := p.SetTempo(120)
	if err != nil {
		t.Errorf("SetTempo(120) unexpected error: %v", err)
	}
	if p.GetBPM() != 120 {
		t.Errorf("SetTempo(120) got BPM %d, want 120", p.GetBPM())
	}

	// Min tempo
	err = p.SetTempo(20)
	if err != nil {
		t.Errorf("SetTempo(20) unexpected error: %v", err)
	}

	// Max tempo
	err = p.SetTempo(300)
	if err != nil {
		t.Errorf("SetTempo(300) unexpected error: %v", err)
	}

	// Too low
	err = p.SetTempo(19)
	if err == nil {
		t.Error("SetTempo(19) should return error")
	}

	// Too high
	err = p.SetTempo(301)
	if err == nil {
		t.Error("SetTempo(301) should return error")
	}
}

// TestDefaultPattern tests that New() creates expected default pattern
func TestDefaultPattern(t *testing.T) {
	p := New()

	// Check default tempo
	if p.BPM != 80 {
		t.Errorf("Default pattern BPM = %d, want 80", p.BPM)
	}

	// Check expected notes (C3, D#3, G3, C3, F3 on steps 1, 4, 5, 9, 13)
	expectedNotes := map[int]uint8{
		0:  48, // Step 1: C3
		3:  51, // Step 4: D#3
		4:  55, // Step 5: G3
		8:  48, // Step 9: C3
		12: 53, // Step 13: F3
	}

	for i := 0; i < NumSteps; i++ {
		if expectedNote, hasNote := expectedNotes[i]; hasNote {
			if p.Steps[i].IsRest {
				t.Errorf("Default pattern step %d should not be rest", i+1)
			}
			if p.Steps[i].Note != expectedNote {
				t.Errorf("Default pattern step %d note = %d, want %d", i+1, p.Steps[i].Note, expectedNote)
			}
		} else {
			if !p.Steps[i].IsRest {
				t.Errorf("Default pattern step %d should be rest", i+1)
			}
		}
	}
}

// TestLoadNonExistent tests loading a pattern that doesn't exist
func TestLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	_, err := Load("nonexistent")
	if err == nil {
		t.Error("Load() of non-existent pattern should return error")
	}
}

// TestList tests listing saved patterns
func TestList(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Initially should be empty (or return empty list if dir doesn't exist)
	patterns, err := List()
	if err != nil {
		t.Errorf("List() unexpected error: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("List() should return empty list initially, got %d patterns", len(patterns))
	}

	// Save a few patterns
	p1 := New()
	p1.Save("pattern_one")

	p2 := New()
	p2.Save("pattern_two")

	// List again
	patterns, err = List()
	if err != nil {
		t.Errorf("List() unexpected error: %v", err)
	}
	if len(patterns) != 2 {
		t.Errorf("List() returned %d patterns, want 2", len(patterns))
	}
}

// TestDelete tests deleting a saved pattern
func TestDelete(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Save a pattern
	p := New()
	p.Save("to_delete")

	// Verify it exists
	patterns, _ := List()
	if len(patterns) != 1 {
		t.Error("Pattern not saved correctly")
	}

	// Delete it
	err := Delete("to_delete")
	if err != nil {
		t.Errorf("Delete() unexpected error: %v", err)
	}

	// Verify it's gone
	patterns, _ = List()
	if len(patterns) != 0 {
		t.Error("Pattern not deleted")
	}

	// Try to delete non-existent
	err = Delete("nonexistent")
	if err == nil {
		t.Error("Delete() of non-existent pattern should return error")
	}
}

// TestSanitizeFilename tests filename sanitization
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple name", "my_pattern", "my_pattern"},
		{"With spaces", "my pattern", "my_pattern"},
		{"Special chars", "my!@#pattern", "mypattern"},
		{"Mixed case", "MyPattern", "MyPattern"},
		{"With hyphens", "my-pattern-1", "my-pattern-1"},
		{"Empty string", "", "unnamed"},
		{"Only special chars", "!@#$", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSaveAndLoad tests saving and loading patterns
func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create a pattern and clear it to start fresh
	p := New()
	p.Clear() // Clear default pattern first
	p.SetNote(1, 60) // C4
	p.SetNote(5, 67) // G4
	p.SetVelocity(1, 120)
	p.SetGate(5, 70)
	p.SetTempo(100)

	// Test ToPatternFile conversion
	pf := p.ToPatternFile("test_pattern")

	if pf.Name != "test_pattern" {
		t.Errorf("PatternFile.Name = %s, want test_pattern", pf.Name)
	}
	if pf.Tempo != 100 {
		t.Errorf("PatternFile.Tempo = %d, want 100", pf.Tempo)
	}
	if len(pf.Steps) != 2 {
		t.Errorf("PatternFile.Steps length = %d, want 2", len(pf.Steps))
	}

	// Test FromPatternFile conversion
	loadedPattern, err := FromPatternFile(pf)
	if err != nil {
		t.Errorf("FromPatternFile() unexpected error: %v", err)
	}

	if loadedPattern.BPM != 100 {
		t.Errorf("Loaded pattern BPM = %d, want 100", loadedPattern.BPM)
	}

	step1, _ := loadedPattern.GetStep(1)
	if step1.Note != 60 || step1.IsRest {
		t.Error("Loaded pattern step 1 not correct")
	}
	if step1.Velocity != 120 {
		t.Errorf("Loaded pattern step 1 velocity = %d, want 120", step1.Velocity)
	}

	step5, _ := loadedPattern.GetStep(5)
	if step5.Note != 67 || step5.IsRest {
		t.Error("Loaded pattern step 5 not correct")
	}
	if step5.Gate != 70 {
		t.Errorf("Loaded pattern step 5 gate = %d, want 70", step5.Gate)
	}

	// Test actual file save/load
	err = p.Save("test_save")
	if err != nil {
		t.Errorf("Save() unexpected error: %v", err)
	}

	// Verify file exists
	expectedFile := filepath.Join(PatternsDir, "test_save.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("Save() did not create file")
	}

	// Load it back
	loadedPattern2, err := Load("test_save")
	if err != nil {
		t.Errorf("Load() unexpected error: %v", err)
	}

	// Verify it matches
	if loadedPattern2.BPM != p.BPM {
		t.Errorf("Loaded pattern BPM = %d, want %d", loadedPattern2.BPM, p.BPM)
	}
}
