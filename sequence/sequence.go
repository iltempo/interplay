package sequence

import (
	"fmt"
	"strings"
	"sync"
)

const (
	NumSteps = 16 // 1 bar = 16 sixteenth notes
)

// Step represents a single step in the sequence
type Step struct {
	Note   uint8 // MIDI note number (0-127), 0 means rest
	IsRest bool  // true if this step is a rest/silence
}

// Pattern represents a 16-step sequence pattern
type Pattern struct {
	Steps [NumSteps]Step
	BPM   int
	mu    sync.RWMutex // protects concurrent access
}

// New creates a new pattern with the default starting pattern:
// Steps 1, 5, 9, 13: C3 (MIDI note 48)
// All other steps: rest
func New() *Pattern {
	p := &Pattern{
		BPM: 80, // default tempo
	}

	// Initialize with default pattern (C3 on beats)
	for i := 0; i < NumSteps; i++ {
		p.Steps[i] = Step{IsRest: true}
	}
	p.Steps[0] = Step{Note: 48, IsRest: false}  // Step 1: C3
	p.Steps[4] = Step{Note: 48, IsRest: false}  // Step 5: C3
	p.Steps[8] = Step{Note: 48, IsRest: false}  // Step 9: C3
	p.Steps[12] = Step{Note: 48, IsRest: false} // Step 13: C3

	return p
}

// SetNote sets a specific step to play a note
// stepNum: 1-16 (user-facing, converts to 0-15 internally)
// note: MIDI note number (0-127)
func (p *Pattern) SetNote(stepNum int, note uint8) error {
	if stepNum < 1 || stepNum > NumSteps {
		return fmt.Errorf("step must be 1-%d", NumSteps)
	}
	if note > 127 {
		return fmt.Errorf("note must be 0-127")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Steps[stepNum-1] = Step{Note: note, IsRest: false}
	return nil
}

// SetRest sets a specific step to be silent
func (p *Pattern) SetRest(stepNum int) error {
	if stepNum < 1 || stepNum > NumSteps {
		return fmt.Errorf("step must be 1-%d", NumSteps)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Steps[stepNum-1] = Step{IsRest: true}
	return nil
}

// Clear resets all steps to rests
func (p *Pattern) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < NumSteps; i++ {
		p.Steps[i] = Step{IsRest: true}
	}
}

// SetTempo changes the BPM
func (p *Pattern) SetTempo(bpm int) error {
	if bpm < 20 || bpm > 300 {
		return fmt.Errorf("BPM must be 20-300")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.BPM = bpm
	return nil
}

// GetStep returns a copy of a specific step (thread-safe read)
func (p *Pattern) GetStep(stepNum int) (Step, error) {
	if stepNum < 1 || stepNum > NumSteps {
		return Step{}, fmt.Errorf("step must be 1-%d", NumSteps)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.Steps[stepNum-1], nil
}

// GetBPM returns the current BPM (thread-safe read)
func (p *Pattern) GetBPM() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.BPM
}

// Clone creates a deep copy of the pattern (for swapping current/next)
func (p *Pattern) Clone() *Pattern {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clone := &Pattern{
		BPM:   p.BPM,
		Steps: p.Steps, // array copy
	}
	return clone
}

// CopyFrom copies the steps and BPM from another pattern (thread-safe)
func (p *Pattern) CopyFrom(other *Pattern) {
	p.mu.Lock()
	defer p.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	p.Steps = other.Steps
	p.BPM = other.BPM
}

// String returns a human-readable representation of the pattern
func (p *Pattern) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tempo: %d BPM\n", p.BPM))
	sb.WriteString("Steps:\n")

	for i := 0; i < NumSteps; i++ {
		step := p.Steps[i]
		stepNum := i + 1
		if step.IsRest {
			sb.WriteString(fmt.Sprintf("  %2d: rest\n", stepNum))
		} else {
			noteName := midiToNoteName(step.Note)
			sb.WriteString(fmt.Sprintf("  %2d: %s (MIDI %d)\n", stepNum, noteName, step.Note))
		}
	}

	return sb.String()
}

// midiToNoteName converts MIDI note number to name (e.g., 60 -> "C4")
func midiToNoteName(note uint8) string {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	octave := int(note/12) - 1
	noteName := noteNames[note%12]
	return fmt.Sprintf("%s%d", noteName, octave)
}

// NoteNameToMIDI converts note name to MIDI number (e.g., "C4" -> 60)
func NoteNameToMIDI(name string) (uint8, error) {
	noteMap := map[string]int{
		"C": 0, "C#": 1, "Db": 1,
		"D": 2, "D#": 3, "Eb": 3,
		"E": 4,
		"F": 5, "F#": 6, "Gb": 6,
		"G": 7, "G#": 8, "Ab": 8,
		"A": 9, "A#": 10, "Bb": 10,
		"B": 11,
	}

	if len(name) < 2 {
		return 0, fmt.Errorf("invalid note name: %s", name)
	}

	// Extract note and octave
	var notePart string
	var octave int

	if len(name) == 2 {
		// e.g., "C4"
		notePart = name[0:1]
		_, err := fmt.Sscanf(name[1:2], "%d", &octave)
		if err != nil {
			return 0, fmt.Errorf("invalid note name: %s", name)
		}
	} else if len(name) == 3 {
		// e.g., "C#4" or "Bb4"
		notePart = name[0:2]
		_, err := fmt.Sscanf(name[2:3], "%d", &octave)
		if err != nil {
			return 0, fmt.Errorf("invalid note name: %s", name)
		}
	} else {
		return 0, fmt.Errorf("invalid note name: %s", name)
	}

	noteValue, ok := noteMap[notePart]
	if !ok {
		return 0, fmt.Errorf("invalid note name: %s", name)
	}

	midiNote := (octave+1)*12 + noteValue
	if midiNote < 0 || midiNote > 127 {
		return 0, fmt.Errorf("note out of range: %s", name)
	}

	return uint8(midiNote), nil
}
