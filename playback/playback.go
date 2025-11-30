package playback

import (
	"fmt"
	"sync"
	"time"

	"github.com/iltempo/interplay/midi"
	"github.com/iltempo/interplay/sequence"
)

// Engine manages the playback loop
type Engine struct {
	midiOut        *midi.Output
	currentPattern *sequence.Pattern
	nextPattern    *sequence.Pattern
	mu             sync.RWMutex
	stopChan       chan struct{}
	stoppedChan    chan struct{}
	verbose        bool
	verboseMu      sync.RWMutex
}

// New creates a new playback engine
func New(midiOut *midi.Output, initialPattern *sequence.Pattern) *Engine {
	return &Engine{
		midiOut:        midiOut,
		currentPattern: initialPattern,
		nextPattern:    initialPattern.Clone(),
		stopChan:       make(chan struct{}),
		stoppedChan:    make(chan struct{}),
	}
}

// GetNextPattern returns the next pattern for modification
func (e *Engine) GetNextPattern() *sequence.Pattern {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.nextPattern
}

// SetVerbose enables or disables step-by-step output
func (e *Engine) SetVerbose(verbose bool) {
	e.verboseMu.Lock()
	defer e.verboseMu.Unlock()
	e.verbose = verbose
}

// IsVerbose returns whether verbose mode is enabled
func (e *Engine) IsVerbose() bool {
	e.verboseMu.RLock()
	defer e.verboseMu.RUnlock()
	return e.verbose
}

// Start begins the playback loop in a goroutine
func (e *Engine) Start() {
	go e.playbackLoop()
}

// Stop stops the playback loop gracefully
func (e *Engine) Stop() {
	close(e.stopChan)
	<-e.stoppedChan // wait for loop to finish
}

// playbackLoop is the main playback goroutine
func (e *Engine) playbackLoop() {
	defer close(e.stoppedChan)

	const channel = 0 // MIDI channel 1 (0-indexed)

	for {
		// Atomically get a clone of the current pattern for this loop iteration.
		// This is the most important part of the concurrency model.
		// The playback loop operates on a completely isolated copy of the pattern.
		e.mu.RLock()
		pattern := e.currentPattern.Clone()
		e.mu.RUnlock()

		bpm := pattern.BPM
		numSteps := len(pattern.Steps)

		// Calculate step duration in milliseconds
		// At 80 BPM: quarter note = 750ms, sixteenth note = 187.5ms
		stepDurationMs := (60_000.0 / float64(bpm)) / 4.0
		stepDuration := time.Duration(stepDurationMs * float64(time.Millisecond))

		// Track active notes with countdown timers
		// map: note number -> remaining steps
		activeNotes := make(map[uint8]int)

		// Play all steps in the pattern
		for stepIdx := 0; stepIdx < numSteps; stepIdx++ {
			// Check for stop signal
			select {
			case <-e.stopChan:
				// Turn off all active notes before stopping
				for note := range activeNotes {
					e.midiOut.NoteOff(channel, note)
				}
				return
			default:
			}

			stepStart := time.Now()

			// Decrement active note counters and send NoteOff if they expire
			for note, stepsRemaining := range activeNotes {
				if stepsRemaining-1 <= 0 {
					err := e.midiOut.NoteOff(channel, note)
					if err != nil {
						fmt.Printf("Error sending Note Off: %v\n", err)
					}
					delete(activeNotes, note)
				} else {
					activeNotes[note] = stepsRemaining - 1
				}
			}

			// Get the current step from our cloned pattern
			step := pattern.Steps[stepIdx]

			if !step.IsRest {
				velocity := step.Velocity
				if velocity == 0 {
					velocity = 100 // default
				}
				gate := step.Gate
				if gate == 0 {
					gate = 90 // default
				}
				duration := step.Duration
				if duration < 1 {
					duration = 1 // default
				}

				// Calculate how many steps the note should sound for, based on gate
				gateSteps := int(float64(duration) * float64(gate) / 100.0)
				if gateSteps < 1 {
					gateSteps = 1 // Note should sound for at least one step
				}

				// If this note is already playing, send a NoteOff first (re-trigger)
				if _, playing := activeNotes[step.Note]; playing {
					err := e.midiOut.NoteOff(channel, step.Note)
					if err != nil {
						fmt.Printf("Error sending Note Off (retrigger): %v\n", err)
					}
					delete(activeNotes, step.Note)
				}

				// Send Note On
				err := e.midiOut.NoteOn(channel, step.Note, velocity)
				if err != nil {
					fmt.Printf("Error sending Note On: %v\n", err)
				}

				if e.IsVerbose() {
					noteName := midiToNoteName(step.Note)
					if duration > 1 {
						fmt.Printf("♪ Step %2d: %s (vel:%d gate:%d%% dur:%d)\n", stepIdx+1, noteName, velocity, gate, duration)
					} else {
						fmt.Printf("♪ Step %2d: %s (vel:%d gate:%d%%)\n", stepIdx+1, noteName, velocity, gate)
					}
				}

				// Add to active notes map to track its duration
				activeNotes[step.Note] = gateSteps
			} else if e.IsVerbose() {
				fmt.Printf("  Step %2d: ---\n", stepIdx+1)
			}

			// Wait for the remainder of the step duration
			elapsed := time.Since(stepStart)
			remaining := stepDuration - elapsed
			if remaining > 0 {
				time.Sleep(remaining)
			}
		}

		// Loop boundary: turn off all remaining active notes (clean cut)
		for note := range activeNotes {
			err := e.midiOut.NoteOff(channel, note)
			if err != nil {
				fmt.Printf("Error sending Note Off (loop boundary): %v\n", err)
			}
		}

		// Swap current ← next. This is the other key part of the concurrency model.
		// We grab the lock, and replace the current pattern with a CLONE of the
		// next pattern. The command handler goroutine can continue to modify the
		// `nextPattern` without interfering with the `currentPattern` that the
		// next loop iteration will use.
		e.mu.Lock()
		e.currentPattern = e.nextPattern.Clone()
		e.mu.Unlock()

		if e.IsVerbose() {
			fmt.Println("--- Loop ---")
		}
	}
}

// midiToNoteName converts MIDI note number to name (same as in sequence package)
func midiToNoteName(note uint8) string {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	octave := int(note/12) - 1
	noteName := noteNames[note%12]
	return fmt.Sprintf("%s%d", noteName, octave)
}
