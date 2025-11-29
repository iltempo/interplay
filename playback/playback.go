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
	const velocity = 100

	for {
		// Get current pattern (snapshot for this loop iteration)
		e.mu.RLock()
		pattern := e.currentPattern
		bpm := pattern.GetBPM()
		e.mu.RUnlock()

		// Calculate step duration in milliseconds
		// At 80 BPM: quarter note = 750ms, sixteenth note = 187.5ms
		stepDurationMs := (60_000.0 / float64(bpm)) / 4.0
		stepDuration := time.Duration(stepDurationMs * float64(time.Millisecond))

		// Gate length: 90% of step duration
		gateDuration := time.Duration(float64(stepDuration) * 0.9)

		// Play all 16 steps
		for stepIdx := 0; stepIdx < sequence.NumSteps; stepIdx++ {
			// Check for stop signal
			select {
			case <-e.stopChan:
				return
			default:
			}

			stepStart := time.Now()
			step := pattern.Steps[stepIdx]

			if !step.IsRest {
				// Send Note On
				err := e.midiOut.NoteOn(channel, step.Note, velocity)
				if err != nil {
					fmt.Printf("Error sending Note On: %v\n", err)
				}

				// Print visual feedback (only if verbose)
				if e.IsVerbose() {
					noteName := midiToNoteName(step.Note)
					fmt.Printf("♪ Step %2d: %s\n", stepIdx+1, noteName)
				}

				// Schedule Note Off after gate duration
				time.AfterFunc(gateDuration, func() {
					err := e.midiOut.NoteOff(channel, step.Note)
					if err != nil {
						fmt.Printf("Error sending Note Off: %v\n", err)
					}
				})
			} else if e.IsVerbose() {
				// Rest - print indicator only if verbose
				fmt.Printf("  Step %2d: ---\n", stepIdx+1)
			}

			// Wait until next step
			elapsed := time.Since(stepStart)
			remaining := stepDuration - elapsed
			if remaining > 0 {
				time.Sleep(remaining)
			}
		}

		// Loop boundary: swap current ← next
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
