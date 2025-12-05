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
	p := New(DefaultPatternLength)

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
	err = p.SetNote(p.Length()+1, 60)
	if err == nil {
		t.Errorf("SetNote(%d, 60) should return error for step out of range", p.Length()+1)
	}

	// Note out of range
	err = p.SetNote(1, 128)
	if err == nil {
		t.Error("SetNote(1, 128) should return error for note > 127")
	}
}

// TestSetVelocity tests velocity setting
func TestSetVelocity(t *testing.T) {
	p := New(DefaultPatternLength)

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
	p := New(DefaultPatternLength)

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
	p := New(DefaultPatternLength)

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
	p := New(DefaultPatternLength)

	// Modify some steps
	p.SetNote(1, 60)
	p.SetNote(5, 67)

	// Clear
	p.Clear()

	// Check all steps are rests
	for i := 0; i < DefaultPatternLength; i++ {
		if !p.Steps[i].IsRest {
			t.Errorf("After Clear(), step %d is not a rest", i+1)
		}
	}
}

// TestClone tests pattern cloning
func TestClone(t *testing.T) {
	p := New(DefaultPatternLength)
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
	p1 := New(DefaultPatternLength)
	p1.SetNote(1, 60)
	p1.SetVelocity(1, 80)
	p1.SetTempo(120)

	p2 := New(DefaultPatternLength)
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

// TestResize tests resizing a pattern
func TestResize(t *testing.T) {
	p := New(8)
	p.SetNote(1, 60)

	// Test expansion
	err := p.Resize(16)
	if err != nil {
		t.Fatalf("Resize(16) unexpected error: %v", err)
	}
	if len(p.Steps) != 16 {
		t.Errorf("Resize(16) length = %d, want 16", len(p.Steps))
	}
	if p.Steps[0].Note != 60 {
		t.Error("Resize(16) did not preserve existing notes")
	}
	for i := 8; i < 16; i++ {
		if !p.Steps[i].IsRest {
			t.Errorf("Resize(16) step %d should be a rest", i+1)
		}
	}

	// Test truncation
	err = p.Resize(4)
	if err != nil {
		t.Fatalf("Resize(4) unexpected error: %v", err)
	}
	if len(p.Steps) != 4 {
		t.Errorf("Resize(4) length = %d, want 4", len(p.Steps))
	}
	if p.Steps[0].Note != 60 {
		t.Error("Resize(4) did not preserve existing notes")
	}

	// Test invalid size
	err = p.Resize(0)
	if err == nil {
		t.Error("Resize(0) should return an error")
	}
	err = p.Resize(-1)
	if err == nil {
		t.Error("Resize(-1) should return an error")
	}
}


// TestSetTempo tests tempo setting
func TestSetTempo(t *testing.T) {
	p := New(DefaultPatternLength)

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
	p := New(DefaultPatternLength)

	// Check default tempo
	if p.BPM != 80 {
		t.Errorf("Default pattern BPM = %d, want 80", p.BPM)
	}

	// Check expected notes in the new bass pattern
	expectedNotes := map[int]struct {
		note     uint8
		duration int
	}{
		0:  {36, 3}, // Step 1:  C2 (long)
		3:  {43, 1}, // Step 4:  G2 (short accent)
		4:  {48, 4}, // Step 5:  C3 (sustained)
		8:  {36, 2}, // Step 9:  C2 (medium)
		10: {39, 1}, // Step 11: D#2 (passing note)
		11: {41, 2}, // Step 12: F2 (medium)
		14: {43, 1}, // Step 15: G2 (staccato)
	}

	for i := 0; i < DefaultPatternLength; i++ {
		if expected, hasNote := expectedNotes[i]; hasNote {
			if p.Steps[i].IsRest {
				t.Errorf("Default pattern step %d should not be rest", i+1)
			}
			if p.Steps[i].Note != expected.note {
				t.Errorf("Default pattern step %d note = %d, want %d", i+1, p.Steps[i].Note, expected.note)
			}
			if p.Steps[i].Duration != expected.duration {
				t.Errorf("Default pattern step %d duration = %d, want %d", i+1, p.Steps[i].Duration, expected.duration)
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
	p1 := New(DefaultPatternLength)
	p1.Save("pattern_one")

	p2 := New(DefaultPatternLength)
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
	p := New(DefaultPatternLength)
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
	p := New(DefaultPatternLength)
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

// TestSetNoteWithDuration tests setting notes with duration
func TestSetNoteWithDuration(t *testing.T) {
	p := New(DefaultPatternLength)

	// Valid duration
	err := p.SetNoteWithDuration(1, 60, 4)
	if err != nil {
		t.Errorf("SetNoteWithDuration(1, 60, 4) unexpected error: %v", err)
	}
	if p.Steps[0].Note != 60 || p.Steps[0].IsRest {
		t.Error("SetNoteWithDuration(1, 60, 4) did not set note correctly")
	}
	if p.Steps[0].Duration != 4 {
		t.Errorf("SetNoteWithDuration(1, 60, 4) duration = %d, want 4", p.Steps[0].Duration)
	}

	// Min duration
	err = p.SetNoteWithDuration(2, 67, 1)
	if err != nil {
		t.Errorf("SetNoteWithDuration(2, 67, 1) unexpected error: %v", err)
	}
	if p.Steps[1].Duration != 1 {
		t.Errorf("SetNoteWithDuration(2, 67, 1) duration = %d, want 1", p.Steps[1].Duration)
	}

	// Max duration
	err = p.SetNoteWithDuration(3, 72, 16)
	if err != nil {
		t.Errorf("SetNoteWithDuration(3, 72, 16) unexpected error: %v", err)
	}
	if p.Steps[2].Duration != 16 {
		t.Errorf("SetNoteWithDuration(3, 72, 16) duration = %d, want 16", p.Steps[2].Duration)
	}

	// Duration too low
	err = p.SetNoteWithDuration(1, 60, 0)
	if err == nil {
		t.Error("SetNoteWithDuration(1, 60, 0) should return error for duration < 1")
	}

	// Duration too high
	err = p.SetNoteWithDuration(1, 60, p.Length()+1)
	if err == nil {
		t.Errorf("SetNoteWithDuration(1, 60, %d) should return error for duration > pattern length", p.Length()+1)
	}

	// Invalid step
	err = p.SetNoteWithDuration(0, 60, 4)
	if err == nil {
		t.Error("SetNoteWithDuration(0, 60, 4) should return error for invalid step")
	}

	// Invalid note
	err = p.SetNoteWithDuration(1, 128, 4)
	if err == nil {
		t.Error("SetNoteWithDuration(1, 128, 4) should return error for invalid note")
	}
}

// TestSetNoteDefaultDuration tests that SetNote uses default duration of 1
func TestSetNoteDefaultDuration(t *testing.T) {
	p := New(DefaultPatternLength)

	err := p.SetNote(1, 60)
	if err != nil {
		t.Errorf("SetNote(1, 60) unexpected error: %v", err)
	}

	if p.Steps[0].Duration != 1 {
		t.Errorf("SetNote should set default duration of 1, got %d", p.Steps[0].Duration)
	}
}

// TestDurationPreservesVelocityGate tests that setting duration preserves existing velocity/gate
func TestDurationPreservesVelocityGate(t *testing.T) {
	p := New(DefaultPatternLength)

	// Set note with custom velocity and gate
	p.SetNote(1, 60)
	p.SetVelocity(1, 120)
	p.SetGate(1, 70)

	// Now change the duration
	err := p.SetNoteWithDuration(1, 60, 4)
	if err != nil {
		t.Errorf("SetNoteWithDuration unexpected error: %v", err)
	}

	// Velocity and gate should be preserved
	if p.Steps[0].Velocity != 120 {
		t.Errorf("Duration change lost velocity, got %d want 120", p.Steps[0].Velocity)
	}
	if p.Steps[0].Gate != 70 {
		t.Errorf("Duration change lost gate, got %d want 70", p.Steps[0].Gate)
	}
	if p.Steps[0].Duration != 4 {
		t.Errorf("Duration not set correctly, got %d want 4", p.Steps[0].Duration)
	}
}

// TestDurationJSONRoundTrip tests that duration is preserved through save/load
func TestDurationJSONRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create pattern with various durations
	p := New(DefaultPatternLength)
	p.Clear()
	p.SetNoteWithDuration(1, 60, 1)  // Default duration
	p.SetNoteWithDuration(5, 67, 4)  // Quarter note
	p.SetNoteWithDuration(9, 72, 8)  // Half note
	p.SetNoteWithDuration(13, 55, 16) // Whole note

	// Save it
	err := p.Save("duration_test")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load("duration_test")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify durations
	tests := []struct {
		step     int
		note     uint8
		duration int
	}{
		{0, 60, 1},
		{4, 67, 4},
		{8, 72, 8},
		{12, 55, 16},
	}

	for _, tt := range tests {
		step := loaded.Steps[tt.step]
		if step.Note != tt.note {
			t.Errorf("Step %d note = %d, want %d", tt.step+1, step.Note, tt.note)
		}
		if step.Duration != tt.duration {
			t.Errorf("Step %d duration = %d, want %d", tt.step+1, step.Duration, tt.duration)
		}
	}
}

// TestDurationInToPatternFile tests that duration is omitted from JSON when default
func TestDurationInToPatternFile(t *testing.T) {
	p := New(DefaultPatternLength)
	p.Clear()

	// Step with default duration (should not appear in JSON)
	p.SetNoteWithDuration(1, 60, 1)

	// Step with non-default duration (should appear in JSON)
	p.SetNoteWithDuration(2, 67, 4)

	pf := p.ToPatternFile("test")

	// Find steps in pattern file
	var step1, step2 *PatternStep
	for i := range pf.Steps {
		if pf.Steps[i].Step == 1 {
			step1 = &pf.Steps[i]
		}
		if pf.Steps[i].Step == 2 {
			step2 = &pf.Steps[i]
		}
	}

	if step1 == nil || step2 == nil {
		t.Fatal("Steps not found in PatternFile")
	}

	// Step 1 should have duration 0 (omitempty, default)
	if step1.Duration != 0 {
		t.Errorf("Step with default duration should have Duration=0 in JSON, got %d", step1.Duration)
	}

	// Step 2 should have duration 4
	if step2.Duration != 4 {
		t.Errorf("Step with duration 4 should have Duration=4 in JSON, got %d", step2.Duration)
	}
}

// TestDurationInClone tests that Clone preserves duration
func TestDurationInClone(t *testing.T) {
	p := New(DefaultPatternLength)
	p.SetNoteWithDuration(1, 60, 8)
	p.SetNoteWithDuration(5, 67, 4)

	clone := p.Clone()

	if clone.Steps[0].Duration != 8 {
		t.Errorf("Clone step 1 duration = %d, want 8", clone.Steps[0].Duration)
	}
	if clone.Steps[4].Duration != 4 {
		t.Errorf("Clone step 5 duration = %d, want 4", clone.Steps[4].Duration)
	}

	// Modify clone and ensure original is unchanged
	clone.SetNoteWithDuration(1, 60, 2)
	if p.Steps[0].Duration != 8 {
		t.Error("Modifying clone duration affected original")
	}
}

// TestDurationInCopyFrom tests that CopyFrom preserves duration
func TestDurationInCopyFrom(t *testing.T) {
	p1 := New(DefaultPatternLength)
	p1.SetNoteWithDuration(1, 60, 8)
	p1.SetNoteWithDuration(5, 67, 4)

	p2 := New(DefaultPatternLength)
	p2.CopyFrom(p1)

	if p2.Steps[0].Duration != 8 {
		t.Errorf("CopyFrom step 1 duration = %d, want 8", p2.Steps[0].Duration)
	}
	if p2.Steps[4].Duration != 4 {
		t.Errorf("CopyFrom step 5 duration = %d, want 4", p2.Steps[4].Duration)
	}
}

// TestDefaultPatternHasDuration tests that default pattern has duration field set correctly
func TestDefaultPatternHasDuration(t *testing.T) {
	p := New(DefaultPatternLength)

	// Just verify that all steps have a valid duration (1-16)
	for i := 0; i < DefaultPatternLength; i++ {
		if !p.Steps[i].IsRest {
			if p.Steps[i].Duration < 1 || p.Steps[i].Duration > 16 {
				t.Errorf("Default pattern step %d duration = %d, should be 1-16", i+1, p.Steps[i].Duration)
			}
		}
	}
}

func TestHumanization(t *testing.T) {
	p := New(16)

	// Test velocity humanization
	err := p.SetHumanizeVelocity(10)
	if err != nil {
		t.Fatalf("SetHumanizeVelocity failed: %v", err)
	}
	if p.Humanization.VelocityRange != 10 {
		t.Errorf("VelocityRange = %d, want 10", p.Humanization.VelocityRange)
	}

	// Test timing humanization
	err = p.SetHumanizeTiming(20)
	if err != nil {
		t.Fatalf("SetHumanizeTiming failed: %v", err)
	}
	if p.Humanization.TimingMs != 20 {
		t.Errorf("TimingMs = %d, want 20", p.Humanization.TimingMs)
	}

	// Test gate humanization
	err = p.SetHumanizeGate(15)
	if err != nil {
		t.Fatalf("SetHumanizeGate failed: %v", err)
	}
	if p.Humanization.GateRange != 15 {
		t.Errorf("GateRange = %d, want 15", p.Humanization.GateRange)
	}

	// Test validation - velocity range
	err = p.SetHumanizeVelocity(100)
	if err == nil {
		t.Error("SetHumanizeVelocity(100) should fail (max is 64)")
	}

	// Test validation - timing range
	err = p.SetHumanizeTiming(100)
	if err == nil {
		t.Error("SetHumanizeTiming(100) should fail (max is 50)")
	}

	// Test validation - gate range
	err = p.SetHumanizeGate(100)
	if err == nil {
		t.Error("SetHumanizeGate(100) should fail (max is 50)")
	}
}

func TestHumanizationInClone(t *testing.T) {
	p := New(16)
	p.SetHumanizeVelocity(10)
	p.SetHumanizeTiming(20)
	p.SetHumanizeGate(15)

	clone := p.Clone()

	if clone.Humanization.VelocityRange != 10 {
		t.Errorf("Cloned VelocityRange = %d, want 10", clone.Humanization.VelocityRange)
	}
	if clone.Humanization.TimingMs != 20 {
		t.Errorf("Cloned TimingMs = %d, want 20", clone.Humanization.TimingMs)
	}
	if clone.Humanization.GateRange != 15 {
		t.Errorf("Cloned GateRange = %d, want 15", clone.Humanization.GateRange)
	}
}

func TestHumanizationInCopyFrom(t *testing.T) {
	p1 := New(16)
	p1.SetHumanizeVelocity(10)
	p1.SetHumanizeTiming(20)
	p1.SetHumanizeGate(15)

	p2 := New(16)
	p2.CopyFrom(p1)

	if p2.Humanization.VelocityRange != 10 {
		t.Errorf("Copied VelocityRange = %d, want 10", p2.Humanization.VelocityRange)
	}
	if p2.Humanization.TimingMs != 20 {
		t.Errorf("Copied TimingMs = %d, want 20", p2.Humanization.TimingMs)
	}
	if p2.Humanization.GateRange != 15 {
		t.Errorf("Copied GateRange = %d, want 15", p2.Humanization.GateRange)
	}
}

func TestDefaultHumanization(t *testing.T) {
	p := New(16)

	// Verify default humanization is set
	if p.Humanization.VelocityRange != 8 {
		t.Errorf("Default VelocityRange = %d, want 8", p.Humanization.VelocityRange)
	}
	if p.Humanization.TimingMs != 10 {
		t.Errorf("Default TimingMs = %d, want 10", p.Humanization.TimingMs)
	}
	if p.Humanization.GateRange != 5 {
		t.Errorf("Default GateRange = %d, want 5", p.Humanization.GateRange)
	}
}

func TestSwing(t *testing.T) {
	p := New(16)

	// Test setting swing
	err := p.SetSwing(50)
	if err != nil {
		t.Fatalf("SetSwing failed: %v", err)
	}
	if p.SwingPercent != 50 {
		t.Errorf("SwingPercent = %d, want 50", p.SwingPercent)
	}

	// Test GetSwing
	swing := p.GetSwing()
	if swing != 50 {
		t.Errorf("GetSwing() = %d, want 50", swing)
	}

	// Test validation - too high
	err = p.SetSwing(100)
	if err == nil {
		t.Error("SetSwing(100) should fail (max is 75)")
	}

	// Test validation - negative
	err = p.SetSwing(-10)
	if err == nil {
		t.Error("SetSwing(-10) should fail")
	}

	// Test setting to 0 (off)
	err = p.SetSwing(0)
	if err != nil {
		t.Fatalf("SetSwing(0) failed: %v", err)
	}
	if p.SwingPercent != 0 {
		t.Errorf("SwingPercent = %d, want 0", p.SwingPercent)
	}
}

func TestSwingInClone(t *testing.T) {
	p := New(16)
	p.SetSwing(50)

	clone := p.Clone()

	if clone.SwingPercent != 50 {
		t.Errorf("Cloned SwingPercent = %d, want 50", clone.SwingPercent)
	}
}

func TestSwingInCopyFrom(t *testing.T) {
	p1 := New(16)
	p1.SetSwing(66)

	p2 := New(16)
	p2.CopyFrom(p1)

	if p2.SwingPercent != 66 {
		t.Errorf("Copied SwingPercent = %d, want 66", p2.SwingPercent)
	}
}

// TestValidateCC tests CC number and value validation
func TestValidateCC(t *testing.T) {
	tests := []struct {
		name      string
		ccNumber  int
		value     int
		wantErr   bool
	}{
		{"Valid min CC", 0, 0, false},
		{"Valid max CC", 127, 127, false},
		{"Valid mid CC", 74, 100, false},
		{"CC number too low", -1, 64, true},
		{"CC number too high", 128, 64, true},
		{"Value too low", 74, -1, true},
		{"Value too high", 74, 128, true},
		{"Both invalid", -1, 128, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCC(tt.ccNumber, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCC(%d, %d) error = %v, wantErr %v", tt.ccNumber, tt.value, err, tt.wantErr)
			}
		})
	}
}

// TestSetGlobalCC tests setting global CC values
func TestSetGlobalCC(t *testing.T) {
	p := New(16)

	// Valid CC
	err := p.SetGlobalCC(74, 127)
	if err != nil {
		t.Errorf("SetGlobalCC(74, 127) unexpected error: %v", err)
	}

	// Verify it was set
	value, ok := p.GetGlobalCC(74)
	if !ok {
		t.Error("SetGlobalCC(74, 127) did not set value")
	}
	if value != 127 {
		t.Errorf("GetGlobalCC(74) = %d, want 127", value)
	}

	// Invalid CC number
	err = p.SetGlobalCC(128, 64)
	if err == nil {
		t.Error("SetGlobalCC(128, 64) should return error for invalid CC number")
	}

	// Invalid value
	err = p.SetGlobalCC(74, 128)
	if err == nil {
		t.Error("SetGlobalCC(74, 128) should return error for invalid value")
	}
}

// TestGetGlobalCC tests getting global CC values
func TestGetGlobalCC(t *testing.T) {
	p := New(16)

	// Get non-existent CC
	value, ok := p.GetGlobalCC(74)
	if ok {
		t.Error("GetGlobalCC should return false for non-existent CC")
	}
	if value != 0 {
		t.Errorf("GetGlobalCC for non-existent CC should return 0, got %d", value)
	}

	// Set and get
	p.SetGlobalCC(74, 100)
	value, ok = p.GetGlobalCC(74)
	if !ok {
		t.Error("GetGlobalCC should return true for existing CC")
	}
	if value != 100 {
		t.Errorf("GetGlobalCC(74) = %d, want 100", value)
	}

	// Set multiple CCs
	p.SetGlobalCC(71, 80)
	p.SetGlobalCC(73, 20)

	value, ok = p.GetGlobalCC(71)
	if !ok || value != 80 {
		t.Errorf("GetGlobalCC(71) = (%d, %v), want (80, true)", value, ok)
	}

	value, ok = p.GetGlobalCC(73)
	if !ok || value != 20 {
		t.Errorf("GetGlobalCC(73) = (%d, %v), want (20, true)", value, ok)
	}
}

// TestGetAllGlobalCC tests getting all global CC values
func TestGetAllGlobalCC(t *testing.T) {
	p := New(16)

	// Initially should be nil
	allCC := p.GetAllGlobalCC()
	if allCC != nil {
		t.Error("GetAllGlobalCC should return nil for new pattern")
	}

	// Set some values
	p.SetGlobalCC(74, 100)
	p.SetGlobalCC(71, 80)

	allCC = p.GetAllGlobalCC()
	if allCC == nil {
		t.Fatal("GetAllGlobalCC should not return nil after setting values")
	}
	if len(allCC) != 2 {
		t.Errorf("GetAllGlobalCC returned %d values, want 2", len(allCC))
	}
	if allCC[74] != 100 {
		t.Errorf("allCC[74] = %d, want 100", allCC[74])
	}
	if allCC[71] != 80 {
		t.Errorf("allCC[71] = %d, want 80", allCC[71])
	}

	// Modifying returned map should not affect pattern
	allCC[74] = 50
	value, _ := p.GetGlobalCC(74)
	if value != 100 {
		t.Error("Modifying GetAllGlobalCC result affected pattern state")
	}
}

// TestGlobalCCInClone tests that Clone copies global CC values
func TestGlobalCCInClone(t *testing.T) {
	p := New(16)
	p.SetGlobalCC(74, 100)
	p.SetGlobalCC(71, 80)

	clone := p.Clone()

	// Verify clone has global CC
	value, ok := clone.GetGlobalCC(74)
	if !ok || value != 100 {
		t.Errorf("Cloned global CC#74 = (%d, %v), want (100, true)", value, ok)
	}

	value, ok = clone.GetGlobalCC(71)
	if !ok || value != 80 {
		t.Errorf("Cloned global CC#71 = (%d, %v), want (80, true)", value, ok)
	}

	// Modify clone and ensure original is unchanged
	clone.SetGlobalCC(74, 50)
	origValue, _ := p.GetGlobalCC(74)
	if origValue != 100 {
		t.Error("Modifying cloned global CC affected original")
	}
}

// TestGlobalCCInCopyFrom tests that CopyFrom copies global CC values
func TestGlobalCCInCopyFrom(t *testing.T) {
	p1 := New(16)
	p1.SetGlobalCC(74, 100)
	p1.SetGlobalCC(71, 80)

	p2 := New(16)
	p2.CopyFrom(p1)

	// Verify p2 has global CC
	value, ok := p2.GetGlobalCC(74)
	if !ok || value != 100 {
		t.Errorf("Copied global CC#74 = (%d, %v), want (100, true)", value, ok)
	}

	value, ok = p2.GetGlobalCC(71)
	if !ok || value != 80 {
		t.Errorf("Copied global CC#71 = (%d, %v), want (80, true)", value, ok)
	}

	// Modify p1 and ensure p2 is unchanged
	p1.SetGlobalCC(74, 50)
	p2Value, _ := p2.GetGlobalCC(74)
	if p2Value != 100 {
		t.Error("Modifying source global CC after CopyFrom affected destination")
	}
}

// TestStepCCValues tests per-step CC automation
func TestStepCCValues(t *testing.T) {
	p := New(16)

	// Set CC on a step
	step := &p.Steps[0]
	if step.CCValues != nil {
		t.Error("New step should have nil CCValues")
	}

	// Manually set CC (simulating what command handlers will do)
	step.CCValues = make(map[int]int)
	step.CCValues[74] = 127
	step.CCValues[71] = 64

	if len(step.CCValues) != 2 {
		t.Errorf("Step CCValues length = %d, want 2", len(step.CCValues))
	}
	if step.CCValues[74] != 127 {
		t.Errorf("Step CC#74 = %d, want 127", step.CCValues[74])
	}
	if step.CCValues[71] != 64 {
		t.Errorf("Step CC#71 = %d, want 64", step.CCValues[71])
	}
}

// TestStepCCInClone tests that Clone deep copies step CC values
func TestStepCCInClone(t *testing.T) {
	p := New(16)
	p.Steps[0].CCValues = map[int]int{74: 127, 71: 64}
	p.Steps[4].CCValues = map[int]int{74: 20}

	clone := p.Clone()

	// Verify clone has CC values
	if len(clone.Steps[0].CCValues) != 2 {
		t.Errorf("Cloned step 1 CCValues length = %d, want 2", len(clone.Steps[0].CCValues))
	}
	if clone.Steps[0].CCValues[74] != 127 {
		t.Errorf("Cloned step 1 CC#74 = %d, want 127", clone.Steps[0].CCValues[74])
	}
	if len(clone.Steps[4].CCValues) != 1 {
		t.Errorf("Cloned step 5 CCValues length = %d, want 1", len(clone.Steps[4].CCValues))
	}

	// Modify clone and ensure original is unchanged
	clone.Steps[0].CCValues[74] = 50
	if p.Steps[0].CCValues[74] != 127 {
		t.Error("Modifying cloned step CC values affected original")
	}

	// Add new CC to clone
	clone.Steps[0].CCValues[73] = 100
	if _, exists := p.Steps[0].CCValues[73]; exists {
		t.Error("Adding CC to cloned step affected original")
	}
}

// TestStepCCInCopyFrom tests that CopyFrom deep copies step CC values
func TestStepCCInCopyFrom(t *testing.T) {
	p1 := New(16)
	p1.Steps[0].CCValues = map[int]int{74: 127, 71: 64}
	p1.Steps[4].CCValues = map[int]int{74: 20}

	p2 := New(16)
	p2.CopyFrom(p1)

	// Verify p2 has CC values
	if len(p2.Steps[0].CCValues) != 2 {
		t.Errorf("Copied step 1 CCValues length = %d, want 2", len(p2.Steps[0].CCValues))
	}
	if p2.Steps[0].CCValues[74] != 127 {
		t.Errorf("Copied step 1 CC#74 = %d, want 127", p2.Steps[0].CCValues[74])
	}

	// Modify p1 and ensure p2 is unchanged
	p1.Steps[0].CCValues[74] = 50
	if p2.Steps[0].CCValues[74] != 127 {
		t.Error("Modifying source step CC values after CopyFrom affected destination")
	}
}

// TestCCJSONRoundTrip tests CC data persistence through save/load
func TestCCJSONRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create pattern with CC automation
	p := New(16)
	p.Clear()
	p.SetNote(1, 60) // C4
	p.SetNote(5, 67) // G4

	// Add CC automation to steps
	p.Steps[0].CCValues = map[int]int{74: 127, 71: 64}
	p.Steps[4].CCValues = map[int]int{74: 20}

	// Save it
	err := p.Save("cc_test")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load("cc_test")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify CC automation was preserved
	if len(loaded.Steps[0].CCValues) != 2 {
		t.Errorf("Loaded step 1 CCValues length = %d, want 2", len(loaded.Steps[0].CCValues))
	}
	if loaded.Steps[0].CCValues[74] != 127 {
		t.Errorf("Loaded step 1 CC#74 = %d, want 127", loaded.Steps[0].CCValues[74])
	}
	if loaded.Steps[0].CCValues[71] != 64 {
		t.Errorf("Loaded step 1 CC#71 = %d, want 64", loaded.Steps[0].CCValues[71])
	}

	if len(loaded.Steps[4].CCValues) != 1 {
		t.Errorf("Loaded step 5 CCValues length = %d, want 1", len(loaded.Steps[4].CCValues))
	}
	if loaded.Steps[4].CCValues[74] != 20 {
		t.Errorf("Loaded step 5 CC#74 = %d, want 20", loaded.Steps[4].CCValues[74])
	}

	// Verify steps without CC have nil CCValues
	if loaded.Steps[1].CCValues != nil {
		t.Error("Steps without CC automation should have nil CCValues after load")
	}
}

// TestCCBackwardCompatibility tests loading patterns without CC data
func TestCCBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create a pattern without CC automation (simulating old format)
	p := New(16)
	p.Clear()
	p.SetNote(1, 60)
	p.SetNote(5, 67)

	// Save it
	err := p.Save("old_format")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load("old_format")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify it loaded successfully
	step1, _ := loaded.GetStep(1)
	if step1.Note != 60 {
		t.Error("Backward compatibility: failed to load note")
	}

	// Verify CC values are nil (not present in old format)
	if loaded.Steps[0].CCValues != nil {
		t.Error("Old format pattern should have nil CCValues")
	}
}

// TestGlobalCCNotPersisted tests that global CC is not saved to JSON
func TestGlobalCCNotPersisted(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create pattern with global CC
	p := New(16)
	p.Clear()
	p.SetNote(1, 60)
	p.SetGlobalCC(74, 100)
	p.SetGlobalCC(71, 80)

	// Save it
	err := p.Save("global_cc_test")
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load("global_cc_test")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify global CC was NOT restored (it's transient)
	allCC := loaded.GetAllGlobalCC()
	if allCC != nil {
		t.Error("Global CC should not be persisted - should be nil after load")
	}

	// Verify note was saved
	step1, _ := loaded.GetStep(1)
	if step1.Note != 60 {
		t.Error("Notes should still be saved correctly")
	}
}

// TestCCOmitEmpty tests that empty CC maps are omitted from JSON
func TestCCOmitEmpty(t *testing.T) {
	p := New(16)
	p.Clear()
	p.SetNote(1, 60)

	// Step 1 has no CC automation
	pf := p.ToPatternFile("test")

	// Find step 1
	var step1 *PatternStep
	for i := range pf.Steps {
		if pf.Steps[i].Step == 1 {
			step1 = &pf.Steps[i]
			break
		}
	}

	if step1 == nil {
		t.Fatal("Step 1 not found in PatternFile")
	}

	// CC field should be nil (omitempty)
	if step1.CC != nil {
		t.Error("Steps without CC automation should have nil CC field in JSON")
	}
}

// TestInvalidCCInJSON tests handling of invalid CC data during load
func TestInvalidCCInJSON(t *testing.T) {
	// This test verifies that invalid CC data in JSON is handled gracefully
	pf := &PatternFile{
		Name:   "test",
		Tempo:  80,
		Length: 16,
		Steps: []PatternStep{
			{
				Step: 1,
				Note: "C4",
				CC: map[string]int{
					"74":  127,  // Valid
					"256": 100,  // Invalid CC number (should be skipped)
					"71":  200,  // Invalid value (should be skipped)
					"abc": 50,   // Invalid CC number format (should be skipped)
				},
			},
		},
	}

	// This should not error, but should skip invalid entries
	p, err := FromPatternFile(pf)
	if err != nil {
		t.Fatalf("FromPatternFile should not error on invalid CC data: %v", err)
	}

	// Only the valid CC#74 should be loaded
	if p.Steps[0].CCValues == nil {
		t.Fatal("Step should have CC values")
	}

	// CC#74 should be present
	if value, ok := p.Steps[0].CCValues[74]; !ok || value != 127 {
		t.Error("Valid CC#74 should be loaded")
	}

	// Invalid entries should not be present (or length should be 1)
	if len(p.Steps[0].CCValues) != 1 {
		t.Errorf("Only valid CC entries should be loaded, got %d entries", len(p.Steps[0].CCValues))
	}
}

// TestSetStepCC tests setting per-step CC automation
func TestSetStepCC(t *testing.T) {
	p := New(16)

	// Valid CC
	err := p.SetStepCC(1, 74, 127)
	if err != nil {
		t.Errorf("SetStepCC(1, 74, 127) unexpected error: %v", err)
	}

	// Verify it was set
	value, ok := p.GetStepCC(1, 74)
	if !ok {
		t.Error("SetStepCC(1, 74, 127) did not set value")
	}
	if value != 127 {
		t.Errorf("GetStepCC(1, 74) = %d, want 127", value)
	}

	// Invalid step
	err = p.SetStepCC(0, 74, 100)
	if err == nil {
		t.Error("SetStepCC(0, 74, 100) should return error for invalid step")
	}

	// Invalid CC number
	err = p.SetStepCC(1, 128, 100)
	if err == nil {
		t.Error("SetStepCC(1, 128, 100) should return error for invalid CC number")
	}

	// Invalid value
	err = p.SetStepCC(1, 74, 128)
	if err == nil {
		t.Error("SetStepCC(1, 74, 128) should return error for invalid value")
	}
}

// TestGetStepCC tests getting per-step CC values
func TestGetStepCC(t *testing.T) {
	p := New(16)

	// Get non-existent CC
	value, ok := p.GetStepCC(1, 74)
	if ok {
		t.Error("GetStepCC should return false for non-existent CC")
	}

	// Set and get
	p.SetStepCC(1, 74, 100)
	value, ok = p.GetStepCC(1, 74)
	if !ok {
		t.Error("GetStepCC should return true for existing CC")
	}
	if value != 100 {
		t.Errorf("GetStepCC(1, 74) = %d, want 100", value)
	}

	// Multiple CCs on same step
	p.SetStepCC(1, 71, 80)
	value, ok = p.GetStepCC(1, 71)
	if !ok || value != 80 {
		t.Errorf("GetStepCC(1, 71) = (%d, %v), want (80, true)", value, ok)
	}

	// Different step
	p.SetStepCC(5, 74, 20)
	value, ok = p.GetStepCC(5, 74)
	if !ok || value != 20 {
		t.Errorf("GetStepCC(5, 74) = (%d, %v), want (20, true)", value, ok)
	}
}

// TestClearStepCC tests clearing CC automation from steps
func TestClearStepCC(t *testing.T) {
	p := New(16)

	// Set multiple CCs on step 1
	p.SetStepCC(1, 74, 127)
	p.SetStepCC(1, 71, 64)

	// Clear specific CC
	err := p.ClearStepCC(1, 74)
	if err != nil {
		t.Errorf("ClearStepCC(1, 74) unexpected error: %v", err)
	}

	// Verify CC#74 is cleared
	_, ok := p.GetStepCC(1, 74)
	if ok {
		t.Error("CC#74 should be cleared from step 1")
	}

	// Verify CC#71 still exists
	value, ok := p.GetStepCC(1, 71)
	if !ok || value != 64 {
		t.Error("CC#71 should still exist on step 1")
	}

	// Clear all CCs from step
	err = p.ClearStepCC(1, -1)
	if err != nil {
		t.Errorf("ClearStepCC(1, -1) unexpected error: %v", err)
	}

	// Verify all CCs are cleared
	_, ok = p.GetStepCC(1, 71)
	if ok {
		t.Error("All CCs should be cleared from step 1")
	}

	// CCValues should be nil after clearing all
	if p.Steps[0].CCValues != nil {
		t.Error("Step CCValues should be nil after clearing all")
	}

	// Clearing from step with no CC should not error
	err = p.ClearStepCC(2, 74)
	if err != nil {
		t.Errorf("ClearStepCC on step with no CC should not error: %v", err)
	}
}

// TestApplyGlobalCC tests converting global CC to per-step automation
func TestApplyGlobalCC(t *testing.T) {
	p := New(16)
	p.Clear()

	// Set some notes
	p.SetNote(1, 60)
	p.SetNote(5, 67)
	p.SetNote(9, 72)

	// Set global CC
	p.SetGlobalCC(74, 100)

	// Apply to all steps with notes
	err := p.ApplyGlobalCC(74)
	if err != nil {
		t.Errorf("ApplyGlobalCC(74) unexpected error: %v", err)
	}

	// Verify steps with notes have CC
	value, ok := p.GetStepCC(1, 74)
	if !ok || value != 100 {
		t.Errorf("Step 1 should have CC#74=100 after apply, got (%d, %v)", value, ok)
	}

	value, ok = p.GetStepCC(5, 74)
	if !ok || value != 100 {
		t.Errorf("Step 5 should have CC#74=100 after apply, got (%d, %v)", value, ok)
	}

	value, ok = p.GetStepCC(9, 74)
	if !ok || value != 100 {
		t.Errorf("Step 9 should have CC#74=100 after apply, got (%d, %v)", value, ok)
	}

	// Verify steps without notes don't have CC
	_, ok = p.GetStepCC(2, 74)
	if ok {
		t.Error("Step 2 (rest) should not have CC after apply")
	}

	// Error if global CC not set
	err = p.ApplyGlobalCC(71)
	if err == nil {
		t.Error("ApplyGlobalCC should error if global CC not set")
	}
}

// TestApplyGlobalCCOverwrite tests that cc-apply overwrites existing values
func TestApplyGlobalCCOverwrite(t *testing.T) {
	p := New(16)
	p.Clear()

	// Set notes with existing CC automation
	p.SetNote(1, 60)
	p.SetStepCC(1, 74, 50) // Existing automation

	// Set different global CC
	p.SetGlobalCC(74, 100)

	// Apply should overwrite
	err := p.ApplyGlobalCC(74)
	if err != nil {
		t.Errorf("ApplyGlobalCC(74) unexpected error: %v", err)
	}

	// Verify value was overwritten
	value, ok := p.GetStepCC(1, 74)
	if !ok || value != 100 {
		t.Errorf("Step 1 CC#74 should be overwritten to 100, got (%d, %v)", value, ok)
	}
}

// TestCCDisplayInString verifies that CC automation is displayed in pattern String() output
func TestCCDisplayInString(t *testing.T) {
	p := New(16)

	// Set notes with CC automation
	p.SetNote(1, 48) // C3
	p.SetStepCC(1, 74, 127)

	p.SetNote(5, 55) // G3
	p.SetStepCC(5, 74, 64)
	p.SetStepCC(5, 71, 100) // Multiple CC on same step

	// Get string representation
	output := p.String()

	// Verify CC indicators appear in output
	if !contains(output, "C3") {
		t.Error("Output should contain note C3")
	}
	if !contains(output, "[CC74:127]") {
		t.Error("Output should contain [CC74:127] indicator for step 1")
	}
	if !contains(output, "G3") {
		t.Error("Output should contain note G3")
	}
	// Step 5 should have both CC values displayed (order may vary due to map iteration)
	if !contains(output, "CC74:64") {
		t.Error("Output should contain CC74:64 indicator for step 5")
	}
	if !contains(output, "CC71:100") {
		t.Error("Output should contain CC71:100 indicator for step 5")
	}

	// Verify rests don't show CC indicators (shouldn't have CC anyway)
	p.SetRest(1)
	output = p.String()
	if contains(output, "[CC74:127]") {
		t.Error("Rest steps should not show CC indicators")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsCheck(s, substr))
}

func containsCheck(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
