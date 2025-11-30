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
