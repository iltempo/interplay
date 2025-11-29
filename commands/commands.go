package commands

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iltempo/interplay/sequence"
)

// VerboseController allows controlling verbose output
type VerboseController interface {
	SetVerbose(bool)
	IsVerbose() bool
}

// Handler processes user commands
type Handler struct {
	pattern           *sequence.Pattern
	verboseController VerboseController
}

// New creates a new command handler
func New(pattern *sequence.Pattern, verboseController VerboseController) *Handler {
	return &Handler{
		pattern:           pattern,
		verboseController: verboseController,
	}
}

// ProcessCommand parses and executes a single command string
func (h *Handler) ProcessCommand(cmdLine string) error {
	cmdLine = strings.TrimSpace(cmdLine)
	if cmdLine == "" {
		// Empty line: show pattern
		return h.handleShow([]string{"show"})
	}

	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "set":
		return h.handleSet(parts)
	case "rest":
		return h.handleRest(parts)
	case "clear":
		return h.handleClear(parts)
	case "tempo":
		return h.handleTempo(parts)
	case "show":
		return h.handleShow(parts)
	case "verbose":
		return h.handleVerbose(parts)
	case "save":
		return h.handleSave(parts)
	case "load":
		return h.handleLoad(parts)
	case "list":
		return h.handleList(parts)
	case "delete":
		return h.handleDelete(parts)
	case "help":
		return h.handleHelp(parts)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// handleSet: set <step> <note>
func (h *Handler) handleSet(parts []string) error {
	if len(parts) != 3 {
		return fmt.Errorf("usage: set <step> <note> (e.g., 'set 1 C4')")
	}

	stepNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	noteName := parts[2]
	midiNote, err := sequence.NoteNameToMIDI(noteName)
	if err != nil {
		return err
	}

	err = h.pattern.SetNote(stepNum, midiNote)
	if err != nil {
		return err
	}

	fmt.Printf("Set step %d to %s (MIDI %d)\n", stepNum, noteName, midiNote)
	return nil
}

// handleRest: rest <step>
func (h *Handler) handleRest(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: rest <step> (e.g., 'rest 1')")
	}

	stepNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	err = h.pattern.SetRest(stepNum)
	if err != nil {
		return err
	}

	fmt.Printf("Set step %d to rest\n", stepNum)
	return nil
}

// handleClear: clear
func (h *Handler) handleClear(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: clear")
	}

	h.pattern.Clear()
	fmt.Println("Cleared all steps")
	return nil
}

// handleTempo: tempo <bpm>
func (h *Handler) handleTempo(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: tempo <bpm> (e.g., 'tempo 120')")
	}

	bpm, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid BPM: %s", parts[1])
	}

	err = h.pattern.SetTempo(bpm)
	if err != nil {
		return err
	}

	fmt.Printf("Set tempo to %d BPM\n", bpm)
	return nil
}

// handleShow: show
func (h *Handler) handleShow(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: show")
	}

	fmt.Println(h.pattern.String())
	return nil
}

// handleVerbose: verbose [on|off]
func (h *Handler) handleVerbose(parts []string) error {
	if len(parts) == 1 {
		// Toggle
		currentState := h.verboseController.IsVerbose()
		h.verboseController.SetVerbose(!currentState)
		if !currentState {
			fmt.Println("Verbose mode enabled (showing steps)")
		} else {
			fmt.Println("Verbose mode disabled")
		}
		return nil
	}

	if len(parts) != 2 {
		return fmt.Errorf("usage: verbose [on|off]")
	}

	switch strings.ToLower(parts[1]) {
	case "on":
		h.verboseController.SetVerbose(true)
		fmt.Println("Verbose mode enabled (showing steps)")
	case "off":
		h.verboseController.SetVerbose(false)
		fmt.Println("Verbose mode disabled")
	default:
		return fmt.Errorf("usage: verbose [on|off]")
	}

	return nil
}

// handleSave: save <name>
func (h *Handler) handleSave(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: save <name> (e.g., 'save my_pattern')")
	}

	// Join remaining parts as the name (allows spaces)
	name := strings.Join(parts[1:], " ")

	err := h.pattern.Save(name)
	if err != nil {
		return fmt.Errorf("failed to save pattern: %w", err)
	}

	fmt.Printf("Saved pattern '%s'\n", name)
	return nil
}

// handleLoad: load <name>
func (h *Handler) handleLoad(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: load <name> (e.g., 'load my_pattern')")
	}

	// Join remaining parts as the name (allows spaces)
	name := strings.Join(parts[1:], " ")

	loadedPattern, err := sequence.Load(name)
	if err != nil {
		return fmt.Errorf("failed to load pattern: %w", err)
	}

	// Copy loaded pattern data into current pattern
	h.pattern.CopyFrom(loadedPattern)

	fmt.Printf("Loaded pattern '%s' (Tempo: %d BPM)\n", name, loadedPattern.BPM)
	return nil
}

// handleList: list
func (h *Handler) handleList(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: list")
	}

	patterns, err := sequence.List()
	if err != nil {
		return fmt.Errorf("failed to list patterns: %w", err)
	}

	if len(patterns) == 0 {
		fmt.Println("No saved patterns found")
		return nil
	}

	fmt.Printf("Saved patterns (%d):\n", len(patterns))
	for _, name := range patterns {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}

// handleDelete: delete <name>
func (h *Handler) handleDelete(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: delete <name> (e.g., 'delete my_pattern')")
	}

	// Join remaining parts as the name (allows spaces)
	name := strings.Join(parts[1:], " ")

	err := sequence.Delete(name)
	if err != nil {
		return fmt.Errorf("failed to delete pattern: %w", err)
	}

	fmt.Printf("Deleted pattern '%s'\n", name)
	return nil
}

// handleHelp: help
func (h *Handler) handleHelp(parts []string) error {
	helpText := `Available commands:
  set <step> <note>   Set a step to play a note (e.g., 'set 1 C4')
  rest <step>         Set a step to rest/silence (e.g., 'rest 1')
  clear               Clear all steps to rests
  tempo <bpm>         Change tempo (e.g., 'tempo 120')
  show                Display current pattern
  verbose [on|off]    Toggle or set verbose step output
  save <name>         Save current pattern (e.g., 'save bass_line')
  load <name>         Load a saved pattern (e.g., 'load bass_line')
  list                List all saved patterns
  delete <name>       Delete a saved pattern (e.g., 'delete bass_line')
  help                Show this help message
  quit                Exit the program
  <enter>             Show current pattern (same as 'show')

Notes can be specified as: C4, D#5, Bb3, etc.
Steps are numbered 1-16.
Patterns are saved in the 'patterns/' directory as JSON files.`

	fmt.Println(helpText)
	return nil
}

// ReadLoop reads commands from input until "quit" or EOF
func (h *Handler) ReadLoop(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(strings.ToLower(line)) == "quit" {
			return nil
		}

		err := h.ProcessCommand(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
