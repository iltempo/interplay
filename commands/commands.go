package commands

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"iltempo.de/midi-ai/sequence"
)

// Handler processes user commands
type Handler struct {
	pattern *sequence.Pattern
}

// New creates a new command handler
func New(pattern *sequence.Pattern) *Handler {
	return &Handler{
		pattern: pattern,
	}
}

// ProcessCommand parses and executes a single command string
func (h *Handler) ProcessCommand(cmdLine string) error {
	cmdLine = strings.TrimSpace(cmdLine)
	if cmdLine == "" {
		return nil
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

// handleHelp: help
func (h *Handler) handleHelp(parts []string) error {
	helpText := `Available commands:
  set <step> <note>   Set a step to play a note (e.g., 'set 1 C4')
  rest <step>         Set a step to rest/silence (e.g., 'rest 1')
  clear               Clear all steps to rests
  tempo <bpm>         Change tempo (e.g., 'tempo 120')
  show                Display current pattern
  help                Show this help message
  quit                Exit the program

Notes can be specified as: C4, D#5, Bb3, etc.
Steps are numbered 1-16.`

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
