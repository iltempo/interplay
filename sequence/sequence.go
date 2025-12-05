package sequence

import (
	"fmt"
	"strings"
	"sync"
)

// DefaultPatternLength is the default length for a new pattern.
const DefaultPatternLength = 16

// Step represents a single step in the sequence
type Step struct {
	Note     uint8       // MIDI note number (0-127), 0 means rest
	IsRest   bool        // true if this step is a rest/silence
	Velocity uint8       // MIDI velocity (0-127), default 100
	Gate     int         // Gate length as percentage (1-100), default 90
	Duration int         // Note duration in steps (1-16), default 1
	CCValues map[int]int // CC automation: CC# → Value (0-127), nil if no automation
}

// Humanization settings for making patterns feel more alive
type Humanization struct {
	VelocityRange int // ± velocity variation (0-64), 0 = off
	TimingMs      int // ± timing variation in milliseconds (0-50), 0 = off
	GateRange     int // ± gate percentage variation (0-50), 0 = off
}

// Pattern represents a musical sequence pattern
type Pattern struct {
	Steps        []Step       // A slice of steps, allowing variable length
	BPM          int
	SwingPercent int          // Swing/groove timing (0-75%), 0 = off, 50 = triplet swing
	Humanization Humanization // humanization settings
	globalCC     map[int]int  // Global CC values (transient, not saved): CC# → Value
	mu           sync.RWMutex // protects concurrent access
}

// New creates a new pattern with a default length and starting sequence.
func New(length int) *Pattern {
	if length <= 0 {
		length = DefaultPatternLength
	}

	p := &Pattern{
		BPM: 80, // default tempo
		// Default humanization for more organic, alive-sounding patterns
		Humanization: Humanization{
			VelocityRange: 8,  // Subtle velocity variation
			TimingMs:      10, // Slight timing variation
			GateRange:     5,  // Small gate variation
		},
		Steps: make([]Step, length),
	}

	// Initialize all steps as rests with defaults
	for i := range p.Steps {
		p.Steps[i] = Step{IsRest: true, Velocity: 100, Gate: 90, Duration: 1}
	}

	// Note: We used to apply a default melodic pattern for 16-step sequences,
	// but starting with silence provides a cleaner slate for users to build upon,
	// especially important for batch/script mode where unexpected sounds are jarring.

	return p
}

// applyDefault16StepPattern populates the sequence with a default melodic bass pattern.
// This is called when a new 16-step pattern is created.
func (p *Pattern) applyDefault16StepPattern() {
	// Melodic bass pattern in C minor with varied durations and dynamics
	// Creates a grooving, atmospheric bass line with long and short notes
	p.Steps[0] = Step{Note: 36, IsRest: false, Velocity: 120, Gate: 85, Duration: 3}  // Step 1:  C2 (long)
	p.Steps[3] = Step{Note: 43, IsRest: false, Velocity: 95, Gate: 60, Duration: 1}   // Step 4:  G2 (short accent)
	p.Steps[4] = Step{Note: 48, IsRest: false, Velocity: 110, Gate: 90, Duration: 4}  // Step 5:  C3 (sustained)
	p.Steps[8] = Step{Note: 36, IsRest: false, Velocity: 115, Gate: 80, Duration: 2}  // Step 9:  C2 (medium)
	p.Steps[10] = Step{Note: 39, IsRest: false, Velocity: 100, Gate: 70, Duration: 1} // Step 11: D#2 (passing note)
	p.Steps[11] = Step{Note: 41, IsRest: false, Velocity: 105, Gate: 75, Duration: 2} // Step 12: F2 (medium)
	p.Steps[14] = Step{Note: 43, IsRest: false, Velocity: 90, Gate: 50, Duration: 1}  // Step 15: G2 (staccato)
}

// SetNoteWithDuration sets a specific step to play a note with a given duration
// stepNum: 1-based (user-facing)
// note: MIDI note number (0-127)
// duration: number of steps the note should last
func (p *Pattern) SetNoteWithDuration(stepNum int, note uint8, duration int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}
	if note > 127 {
		return fmt.Errorf("note must be 0-127")
	}
	if duration < 1 || duration > numSteps {
		return fmt.Errorf("duration must be 1-%d steps", numSteps)
	}

	// Preserve existing velocity/gate if step already has values, otherwise use defaults
	existingStep := p.Steps[stepNum-1]
	velocity := existingStep.Velocity
	gate := existingStep.Gate
	if velocity == 0 {
		velocity = 100
	}
	if gate == 0 {
		gate = 90
	}

	p.Steps[stepNum-1] = Step{
		Note:     note,
		IsRest:   false,
		Velocity: velocity,
		Gate:     gate,
		Duration: duration,
	}
	return nil
}

// SetNote sets a specific step to play a note with default duration (1 step)
// stepNum: 1-based (user-facing)
// note: MIDI note number (0-127)
func (p *Pattern) SetNote(stepNum int, note uint8) error {
	return p.SetNoteWithDuration(stepNum, note, 1)
}

// SetRest sets a specific step to be silent
func (p *Pattern) SetRest(stepNum int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}

	p.Steps[stepNum-1] = Step{IsRest: true, Velocity: 100, Gate: 90, Duration: 1}
	return nil
}

// SetVelocity sets the velocity for a specific step
func (p *Pattern) SetVelocity(stepNum int, velocity uint8) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}
	if velocity > 127 {
		return fmt.Errorf("velocity must be 0-127")
	}

	p.Steps[stepNum-1].Velocity = velocity
	return nil
}

// SetGate sets the gate length (as percentage) for a specific step
func (p *Pattern) SetGate(stepNum int, gate int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}
	if gate < 1 || gate > 100 {
		return fmt.Errorf("gate must be 1-100 (percentage)")
	}

	p.Steps[stepNum-1].Gate = gate
	return nil
}

// Clear resets all steps to rests
func (p *Pattern) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.Steps {
		p.Steps[i] = Step{IsRest: true, Velocity: 100, Gate: 90, Duration: 1}
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
	p.mu.RLock()
	defer p.mu.RUnlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return Step{}, fmt.Errorf("step must be 1-%d", numSteps)
	}

	return p.Steps[stepNum-1], nil
}

// GetBPM returns the current BPM (thread-safe read)
func (p *Pattern) GetBPM() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.BPM
}

// Length returns the number of steps in the pattern (thread-safe).
func (p *Pattern) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.Steps)
}

// Clone creates a deep copy of the pattern (for swapping current/next)
func (p *Pattern) Clone() *Pattern {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clone := &Pattern{
		BPM:          p.BPM,
		SwingPercent: p.SwingPercent,
		Humanization: p.Humanization, // Copy humanization settings
		Steps:        make([]Step, len(p.Steps)),
	}

	// Deep copy steps (including CC values maps)
	for i, step := range p.Steps {
		clone.Steps[i] = step
		// Deep copy CCValues map if present
		if step.CCValues != nil {
			clone.Steps[i].CCValues = make(map[int]int)
			for ccNum, value := range step.CCValues {
				clone.Steps[i].CCValues[ccNum] = value
			}
		}
	}

	// Deep copy globalCC map if present
	if p.globalCC != nil {
		clone.globalCC = make(map[int]int)
		for ccNum, value := range p.globalCC {
			clone.globalCC[ccNum] = value
		}
	}

	return clone
}

// CopyFrom copies the steps and BPM from another pattern (thread-safe)
func (p *Pattern) CopyFrom(other *Pattern) {
	p.mu.Lock()
	defer p.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	p.BPM = other.BPM
	p.SwingPercent = other.SwingPercent
	p.Humanization = other.Humanization

	// Deep copy steps (including CC values)
	p.Steps = make([]Step, len(other.Steps))
	for i, step := range other.Steps {
		p.Steps[i] = step
		if step.CCValues != nil {
			p.Steps[i].CCValues = make(map[int]int)
			for ccNum, value := range step.CCValues {
				p.Steps[i].CCValues[ccNum] = value
			}
		}
	}

	// Deep copy globalCC
	if other.globalCC != nil {
		p.globalCC = make(map[int]int)
		for ccNum, value := range other.globalCC {
			p.globalCC[ccNum] = value
		}
	} else {
		p.globalCC = nil
	}
}

// Resize changes the number of steps in the pattern.
// If the new length is greater, new steps are added as rests.
// If the new length is smaller, the pattern is truncated.
func (p *Pattern) Resize(newLength int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if newLength <= 0 {
		return fmt.Errorf("length must be positive")
	}

	currentLength := len(p.Steps)
	if newLength == currentLength {
		return nil // No change needed
	}

	newSteps := make([]Step, newLength)

	// Copy existing steps
	copy(newSteps, p.Steps)

	// If expanding, initialize new steps as rests
	if newLength > currentLength {
		for i := currentLength; i < newLength; i++ {
			newSteps[i] = Step{IsRest: true, Velocity: 100, Gate: 90, Duration: 1}
		}
	}

	p.Steps = newSteps
	return nil
}

// String returns a human-readable representation of the pattern
func (p *Pattern) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tempo: %d BPM, Length: %d steps\n", p.BPM, len(p.Steps)))
	sb.WriteString("Steps:\n")

	for i, step := range p.Steps {
		stepNum := i + 1
		if step.IsRest {
			sb.WriteString(fmt.Sprintf("  %2d: rest\n", stepNum))
		} else {
			noteName := midiToNoteName(step.Note)
			// Build base info string
			var info string
			if step.Duration > 1 {
				info = fmt.Sprintf("  %2d: %s (vel:%d gate:%d%% dur:%d)", stepNum, noteName, step.Velocity, step.Gate, step.Duration)
			} else {
				info = fmt.Sprintf("  %2d: %s (vel:%d gate:%d%%)", stepNum, noteName, step.Velocity, step.Gate)
			}

			// Add CC automation indicators if present
			if len(step.CCValues) > 0 {
				info += " ["
				first := true
				for ccNum, value := range step.CCValues {
					if !first {
						info += ", "
					}
					info += fmt.Sprintf("CC%d:%d", ccNum, value)
					first = false
				}
				info += "]"
			}

			sb.WriteString(info + "\n")
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

// SetHumanizeVelocity sets the velocity humanization range (0-64)
func (p *Pattern) SetHumanizeVelocity(amount int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if amount < 0 || amount > 64 {
		return fmt.Errorf("velocity humanization must be 0-64")
	}
	p.Humanization.VelocityRange = amount
	return nil
}

// SetHumanizeTiming sets the timing humanization in milliseconds (0-50)
func (p *Pattern) SetHumanizeTiming(ms int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ms < 0 || ms > 50 {
		return fmt.Errorf("timing humanization must be 0-50ms")
	}
	p.Humanization.TimingMs = ms
	return nil
}

// SetHumanizeGate sets the gate humanization range (0-50)
func (p *Pattern) SetHumanizeGate(amount int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if amount < 0 || amount > 50 {
		return fmt.Errorf("gate humanization must be 0-50")
	}
	p.Humanization.GateRange = amount
	return nil
}

// GetHumanization returns a copy of the current humanization settings
func (p *Pattern) GetHumanization() Humanization {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Humanization
}

// SetSwing sets the swing/groove percentage (0-75%)
// 0 = straight timing, 50 = triplet swing, 66 = hard swing
func (p *Pattern) SetSwing(percent int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if percent < 0 || percent > 75 {
		return fmt.Errorf("swing must be 0-75%%")
	}
	p.SwingPercent = percent
	return nil
}

// GetSwing returns the current swing percentage
func (p *Pattern) GetSwing() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.SwingPercent
}

// SetGlobalCC sets a global CC value (transient, not saved with pattern)
// Global CC values are sent at the start of each loop iteration
func (p *Pattern) SetGlobalCC(ccNumber, value int) error {
	if err := ValidateCC(ccNumber, value); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.globalCC == nil {
		p.globalCC = make(map[int]int)
	}
	p.globalCC[ccNumber] = value
	return nil
}

// GetGlobalCC returns the global CC value for a specific CC number
// Returns (value, true) if set, (0, false) if not set
func (p *Pattern) GetGlobalCC(ccNumber int) (int, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.globalCC == nil {
		return 0, false
	}
	value, ok := p.globalCC[ccNumber]
	return value, ok
}

// GetAllGlobalCC returns a copy of all global CC values
func (p *Pattern) GetAllGlobalCC() map[int]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.globalCC == nil {
		return nil
	}

	// Return a copy to prevent external modification
	copy := make(map[int]int)
	for ccNum, value := range p.globalCC {
		copy[ccNum] = value
	}
	return copy
}

// SetStepCC sets a CC value for a specific step
func (p *Pattern) SetStepCC(stepNum, ccNumber, value int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}

	if err := ValidateCC(ccNumber, value); err != nil {
		return err
	}

	step := &p.Steps[stepNum-1]
	if step.CCValues == nil {
		step.CCValues = make(map[int]int)
	}
	step.CCValues[ccNumber] = value
	return nil
}

// GetStepCC returns the CC value for a specific step and CC number
// Returns (value, true) if set, (0, false) if not set
func (p *Pattern) GetStepCC(stepNum, ccNumber int) (int, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return 0, false
	}

	step := &p.Steps[stepNum-1]
	if step.CCValues == nil {
		return 0, false
	}

	value, ok := step.CCValues[ccNumber]
	return value, ok
}

// ClearStepCC removes a specific CC automation from a step
// If ccNumber is -1, clears all CC automation from the step
func (p *Pattern) ClearStepCC(stepNum, ccNumber int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	numSteps := len(p.Steps)
	if stepNum < 1 || stepNum > numSteps {
		return fmt.Errorf("step must be 1-%d", numSteps)
	}

	step := &p.Steps[stepNum-1]
	if step.CCValues == nil {
		return nil // Nothing to clear
	}

	if ccNumber == -1 {
		// Clear all CC automation
		step.CCValues = nil
	} else {
		// Clear specific CC number
		delete(step.CCValues, ccNumber)
		// If map is now empty, set to nil
		if len(step.CCValues) == 0 {
			step.CCValues = nil
		}
	}

	return nil
}

// ApplyGlobalCC applies a global CC value to all steps with notes
func (p *Pattern) ApplyGlobalCC(ccNumber int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if global CC is set
	if p.globalCC == nil {
		return fmt.Errorf("no global CC values set")
	}

	value, ok := p.globalCC[ccNumber]
	if !ok {
		return fmt.Errorf("no global value set for CC#%d", ccNumber)
	}

	// Apply to all steps with notes
	for i := range p.Steps {
		if !p.Steps[i].IsRest {
			if p.Steps[i].CCValues == nil {
				p.Steps[i].CCValues = make(map[int]int)
			}
			p.Steps[i].CCValues[ccNumber] = value
		}
	}

	return nil
}
