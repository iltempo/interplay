package commands

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/iltempo/interplay/ai"
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
	aiClient          *ai.Client
}

// New creates a new command handler
func New(pattern *sequence.Pattern, verboseController VerboseController) *Handler {
	// Try to initialize AI client (optional)
	aiClient, _ := ai.NewFromEnv()

	return &Handler{
		pattern:           pattern,
		verboseController: verboseController,
		aiClient:          aiClient,
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
	case "reset":
		return h.handleReset(parts)
	case "tempo":
		return h.handleTempo(parts)
	case "velocity":
		return h.handleVelocity(parts)
	case "gate":
		return h.handleGate(parts)
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
	case "ai":
		return h.handleAI(parts)
	case "ask":
		return h.handleAsk(cmdLine)
	case "clear-chat":
		return h.handleClearChat(parts)
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

// handleReset: reset
func (h *Handler) handleReset(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: reset")
	}

	// Create a new default pattern
	defaultPattern := sequence.New()

	// Copy it into the current pattern
	h.pattern.CopyFrom(defaultPattern)

	fmt.Println("Reset to default pattern")
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

// handleVelocity: velocity <step> <value>
func (h *Handler) handleVelocity(parts []string) error {
	if len(parts) != 3 {
		return fmt.Errorf("usage: velocity <step> <value> (e.g., 'velocity 1 80')")
	}

	stepNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	velocity, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid velocity: %s", parts[2])
	}

	if velocity < 0 || velocity > 127 {
		return fmt.Errorf("velocity must be 0-127")
	}

	err = h.pattern.SetVelocity(stepNum, uint8(velocity))
	if err != nil {
		return err
	}

	fmt.Printf("Set step %d velocity to %d\n", stepNum, velocity)
	return nil
}

// handleGate: gate <step> <percentage>
func (h *Handler) handleGate(parts []string) error {
	if len(parts) != 3 {
		return fmt.Errorf("usage: gate <step> <percentage> (e.g., 'gate 1 50')")
	}

	stepNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	gate, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid gate: %s", parts[2])
	}

	err = h.pattern.SetGate(stepNum, gate)
	if err != nil {
		return err
	}

	fmt.Printf("Set step %d gate to %d%%\n", stepNum, gate)
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

// handleAI: ai - enter interactive AI session
func (h *Handler) handleAI(parts []string) error {
	// Check if AI client is available
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	if len(parts) != 1 {
		return fmt.Errorf("usage: ai (enter interactive session)")
	}

	// Clear any previous conversation history to start fresh
	h.aiClient.ClearHistory()

	fmt.Println("Entering AI session. Type 'exit' to return to command mode.")
	fmt.Println()

	// Create readline for AI session
	rl, err := readline.New("AI> ")
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	ctx := context.Background()

	for {
		// Read user input
		input, err := rl.Readline()
		if err != nil { // io.EOF or other error
			fmt.Println("\nExiting AI session.")
			return nil
		}

		input = strings.TrimSpace(input)

		// Check for exit command
		if strings.ToLower(input) == "exit" {
			fmt.Println("Exiting AI session.")
			return nil
		}

		if input == "" {
			continue
		}

		// Get current pattern state
		currentPattern := h.pattern.String()

		// Send to AI
		response, err := h.aiClient.Session(ctx, input, currentPattern)
		if err != nil {
			fmt.Printf("AI error: %v\n", err)
			continue
		}

		// Print AI response (clean up [EXECUTE] blocks for display)
		displayMessage := cleanExecuteBlocks(response.Message)
		fmt.Printf("\n%s\n", displayMessage)

		// Execute any commands
		if len(response.Commands) > 0 {
			fmt.Printf("\nExecuting %d command(s):\n", len(response.Commands))
			for _, cmd := range response.Commands {
				fmt.Printf("  > %s\n", cmd)
				if err := h.ProcessCommand(cmd); err != nil {
					fmt.Printf("  Error: %v\n", err)
				}
			}
		}

		fmt.Println()
	}
}

// cleanExecuteBlocks removes [EXECUTE]...[/EXECUTE] blocks from display
func cleanExecuteBlocks(text string) string {
	executeStart := "[EXECUTE]"
	executeEnd := "[/EXECUTE]"

	for {
		startIdx := strings.Index(text, executeStart)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(text[startIdx:], executeEnd)
		if endIdx == -1 {
			break
		}

		// Remove the entire block including markers
		text = text[:startIdx] + text[startIdx+endIdx+len(executeEnd):]
	}

	return strings.TrimSpace(text)
}

// handleAsk: ask <question>
func (h *Handler) handleAsk(cmdLine string) error {
	// Check if AI client is available
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	// Extract the question (everything after "ask ")
	question := strings.TrimSpace(strings.TrimPrefix(cmdLine, "ask"))
	if question == "" {
		return fmt.Errorf("usage: ask <question> (e.g., 'ask what scale is this in?')")
	}

	// Get current pattern state
	currentPattern := h.pattern.String()

	fmt.Println("AI thinking...")

	// Get conversational response
	ctx := context.Background()
	response, err := h.aiClient.Chat(ctx, question, currentPattern)
	if err != nil {
		return fmt.Errorf("AI error: %w", err)
	}

	// Print response
	fmt.Printf("\n%s\n\n", response)

	return nil
}

// handleClearChat: clear-chat
func (h *Handler) handleClearChat(parts []string) error {
	// Check if AI client is available
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	if len(parts) != 1 {
		return fmt.Errorf("usage: clear-chat")
	}

	h.aiClient.ClearHistory()
	fmt.Println("Conversation history cleared")
	return nil
}

// handleHelp: help
func (h *Handler) handleHelp(parts []string) error {
	aiStatus := "disabled"
	if h.aiClient != nil {
		aiStatus = "enabled"
	}

	helpText := fmt.Sprintf(`Available commands:
  set <step> <note>       Set a step to play a note (e.g., 'set 1 C4')
  rest <step>             Set a step to rest/silence (e.g., 'rest 1')
  velocity <step> <val>   Set step velocity 0-127 (e.g., 'velocity 1 80')
  gate <step> <percent>   Set step gate length 1-100%% (e.g., 'gate 1 50')
  clear                   Clear all steps to rests
  reset                   Reset to default pattern
  tempo <bpm>             Change tempo (e.g., 'tempo 120')
  show                    Display current pattern
  verbose [on|off]        Toggle or set verbose step output
  save <name>             Save current pattern (e.g., 'save bass_line')
  load <name>             Load a saved pattern (e.g., 'load bass_line')
  list                    List all saved patterns
  delete <name>           Delete a saved pattern (e.g., 'delete bass_line')
  ai                      Enter interactive AI session (AI: %s)
                          Conversational mode - ask questions, get suggestions,
                          and apply changes in real-time. Type 'exit' to return.
  ask <question>          Ask AI a single question (AI: %s)
                          Quick questions without entering session mode
  clear-chat              Clear AI conversation history
  help                    Show this help message
  quit                    Exit the program
  <enter>                 Show current pattern (same as 'show')

Notes: C4, D#5, Bb3, etc. | Steps: 1-16 | Default velocity: 100 | Default gate: 90%%
Patterns saved in 'patterns/' directory as JSON files.
AI features require ANTHROPIC_API_KEY environment variable.`, aiStatus, aiStatus)

	fmt.Println(helpText)
	return nil
}

// ReadLoop reads commands from input until "quit" or EOF
func (h *Handler) ReadLoop(reader io.Reader) error {
	// Configure readline with history
	rl, err := readline.New("> ")
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF or other error
			return nil
		}

		if strings.TrimSpace(strings.ToLower(line)) == "quit" {
			return nil
		}

		err = h.ProcessCommand(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
