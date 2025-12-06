package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/chzyer/readline"
	"github.com/iltempo/interplay/ai"
	"github.com/iltempo/interplay/comparison"
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
	currentModel      string // Current model ID (e.g., "haiku", "sonnet", "opus")
}

// New creates a new command handler with the default AI model
func New(pattern *sequence.Pattern, verboseController VerboseController) *Handler {
	return NewWithModel(pattern, verboseController, "")
}

// NewWithModel creates a new command handler with a specific AI model
// If model is empty, uses the default model (Haiku)
func NewWithModel(pattern *sequence.Pattern, verboseController VerboseController, model anthropic.Model) *Handler {
	// Try to initialize AI client (optional)
	var aiClient *ai.Client
	var err error

	// Determine current model ID
	currentModelID := "haiku" // default
	if model != "" {
		// Find the model ID from the API model string
		for _, m := range comparison.AvailableModels {
			if string(m.APIModel) == string(model) {
				currentModelID = m.ID
				break
			}
		}
		aiClient, err = ai.NewFromEnvWithModel(model)
	} else {
		aiClient, err = ai.NewFromEnv()
	}

	// Ignore error - AI is optional
	_ = err

	return &Handler{
		pattern:           pattern,
		verboseController: verboseController,
		aiClient:          aiClient,
		currentModel:      currentModelID,
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
	case "humanize":
		return h.handleHumanize(parts)
	case "swing":
		return h.handleSwing(parts)
	case "length":
		return h.handleLength(parts)
	case "cc":
		return h.handleCC(parts)
	case "cc-step":
		return h.handleCCStep(parts)
	case "cc-clear":
		return h.handleCCClear(parts)
	case "cc-apply":
		return h.handleCCApply(parts)
	case "cc-show":
		return h.handleCCShow(parts)
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
	case "clear-chat":
		return h.handleClearChat(parts)
	case "compare":
		return h.handleCompare(parts)
	case "compare-list":
		return h.handleCompareList(parts)
	case "compare-view":
		return h.handleCompareView(parts)
	case "compare-load":
		return h.handleCompareLoad(parts)
	case "compare-delete":
		return h.handleCompareDelete(parts)
	case "models":
		return h.handleModels(parts)
	case "model":
		return h.handleModel(parts)
	case "blind":
		return h.handleBlind(parts)
	case "compare-rate":
		return h.handleCompareRate(parts)
	case "help":
		return h.handleHelp(parts)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// handleSet: set <step> <note|rest> [vel:<value>] [gate:<percent>] [dur:<steps>]
func (h *Handler) handleSet(parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("usage: set <step> <note|rest> [vel:<value>] [gate:<percent>] [dur:<steps>]\n" +
			"e.g., 'set 1 C4' or 'set 1 rest' or 'set 1 C4 vel:120 gate:85 dur:3'")
	}

	stepNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	noteName := parts[2]

	// Check if user wants to set a rest
	if strings.ToLower(noteName) == "rest" {
		err := h.pattern.SetRest(stepNum)
		if err != nil {
			return err
		}
		fmt.Printf("Set step %d to rest\n", stepNum)
		return nil
	}

	midiNote, err := sequence.NoteNameToMIDI(noteName)
	if err != nil {
		return err
	}

	// Get pattern length for validation
	patternLen := h.pattern.Length()

	// Parse optional parameters
	var velocity *uint8
	var gate *int
	duration := 1 // default

	for i := 3; i < len(parts); i++ {
		param := parts[i]

		if strings.HasPrefix(param, "vel:") {
			velStr := strings.TrimPrefix(param, "vel:")
			velInt, err := strconv.Atoi(velStr)
			if err != nil {
				return fmt.Errorf("invalid velocity: %s", velStr)
			}
			if velInt < 0 || velInt > 127 {
				return fmt.Errorf("velocity must be 0-127, got %d", velInt)
			}
			vel := uint8(velInt)
			velocity = &vel

		} else if strings.HasPrefix(param, "gate:") {
			gateStr := strings.TrimPrefix(param, "gate:")
			gateInt, err := strconv.Atoi(gateStr)
			if err != nil {
				return fmt.Errorf("invalid gate: %s", gateStr)
			}
			if gateInt < 1 || gateInt > 100 {
				return fmt.Errorf("gate must be 1-100%%, got %d", gateInt)
			}
			gate = &gateInt

		} else if strings.HasPrefix(param, "dur:") {
			durStr := strings.TrimPrefix(param, "dur:")
			duration, err = strconv.Atoi(durStr)
			if err != nil {
				return fmt.Errorf("invalid duration: %s", durStr)
			}
			if duration < 1 || duration > patternLen {
				return fmt.Errorf("duration must be 1-%d steps", patternLen)
			}

		} else {
			return fmt.Errorf("unknown parameter: %s (expected vel:, gate:, or dur:)", param)
		}
	}

	// Set the note with duration
	err = h.pattern.SetNoteWithDuration(stepNum, midiNote, duration)
	if err != nil {
		return err
	}

	// Apply velocity if specified
	if velocity != nil {
		err = h.pattern.SetVelocity(stepNum, *velocity)
		if err != nil {
			return err
		}
	}

	// Apply gate if specified
	if gate != nil {
		err = h.pattern.SetGate(stepNum, *gate)
		if err != nil {
			return err
		}
	}

	// Build output message
	msg := fmt.Sprintf("Set step %d to %s (MIDI %d", stepNum, noteName, midiNote)
	if velocity != nil {
		msg += fmt.Sprintf(", vel:%d", *velocity)
	}
	if gate != nil {
		msg += fmt.Sprintf(", gate:%d%%", *gate)
	}
	if duration > 1 {
		msg += fmt.Sprintf(", dur:%d", duration)
	}
	msg += ")"
	fmt.Println(msg)

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

	// Create a new default pattern (with default length)
	defaultPattern := sequence.New(sequence.DefaultPatternLength)

	// Copy it into the current pattern
	h.pattern.CopyFrom(defaultPattern)

	fmt.Printf("Reset to default %d-step pattern\n", sequence.DefaultPatternLength)
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

// handleHumanize: humanize <type> <amount>
// type: velocity, timing, gate
// amount: 0-64 for velocity, 0-50 for timing (ms), 0-50 for gate
func (h *Handler) handleHumanize(parts []string) error {
	if len(parts) == 1 {
		// Show current humanization settings
		humanization := h.pattern.GetHumanization()
		fmt.Printf("Humanization settings:\n")
		fmt.Printf("  velocity: ±%d (0-64)\n", humanization.VelocityRange)
		fmt.Printf("  timing:   ±%dms (0-50)\n", humanization.TimingMs)
		fmt.Printf("  gate:     ±%d%% (0-50)\n", humanization.GateRange)
		if humanization.VelocityRange == 0 && humanization.TimingMs == 0 && humanization.GateRange == 0 {
			fmt.Println("  (humanization is OFF)")
		}
		return nil
	}

	if len(parts) != 3 {
		return fmt.Errorf("usage: humanize <type> <amount> (e.g., 'humanize velocity 10')\n" +
			"types: velocity (0-64), timing (0-50ms), gate (0-50)\n" +
			"or: humanize (to show current settings)")
	}

	humanizeType := strings.ToLower(parts[1])
	amount, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid amount: %s", parts[2])
	}

	switch humanizeType {
	case "velocity", "vel":
		err = h.pattern.SetHumanizeVelocity(amount)
		if err != nil {
			return err
		}
		if amount == 0 {
			fmt.Println("Velocity humanization OFF")
		} else {
			fmt.Printf("Velocity humanization set to ±%d\n", amount)
		}

	case "timing", "time":
		err = h.pattern.SetHumanizeTiming(amount)
		if err != nil {
			return err
		}
		if amount == 0 {
			fmt.Println("Timing humanization OFF")
		} else {
			fmt.Printf("Timing humanization set to ±%dms\n", amount)
		}

	case "gate":
		err = h.pattern.SetHumanizeGate(amount)
		if err != nil {
			return err
		}
		if amount == 0 {
			fmt.Println("Gate humanization OFF")
		} else {
			fmt.Printf("Gate humanization set to ±%d%%\n", amount)
		}

	default:
		return fmt.Errorf("unknown humanize type: %s (use: velocity, timing, or gate)", humanizeType)
	}

	return nil
}

// handleSwing: swing <percent>
// percent: 0-75%, where 0 = straight, 50 = triplet swing, 66 = hard swing
func (h *Handler) handleSwing(parts []string) error {
	if len(parts) == 1 {
		// Show current swing setting
		swing := h.pattern.GetSwing()
		if swing == 0 {
			fmt.Println("Swing: OFF (straight timing)")
		} else {
			fmt.Printf("Swing: %d%%", swing)
			if swing >= 48 && swing <= 52 {
				fmt.Println(" (triplet swing)")
			} else if swing >= 64 && swing <= 68 {
				fmt.Println(" (hard swing)")
			} else {
				fmt.Println()
			}
		}
		return nil
	}

	if len(parts) != 2 {
		return fmt.Errorf("usage: swing <percent> (e.g., 'swing 50' for triplet swing)\n" +
			"0 = straight, 50 = triplet swing, 66 = hard swing\n" +
			"or: swing (to show current setting)")
	}

	percent, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid swing percentage: %s", parts[1])
	}

	err = h.pattern.SetSwing(percent)
	if err != nil {
		return err
	}

	if percent == 0 {
		fmt.Println("Swing OFF - straight timing")
	} else {
		fmt.Printf("Swing set to %d%%", percent)
		if percent >= 48 && percent <= 52 {
			fmt.Println(" (triplet swing - classic feel)")
		} else if percent >= 64 && percent <= 68 {
			fmt.Println(" (hard swing - laid back groove)")
		} else {
			fmt.Println()
		}
	}

	return nil
}

// handleLength: length <steps>
func (h *Handler) handleLength(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: length <steps> (e.g., 'length 32')")
	}

	length, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid length: %s", parts[1])
	}

	err = h.pattern.Resize(length)
	if err != nil {
		return err
	}

	fmt.Printf("Pattern length set to %d steps\n", length)
	return nil
}

// handleSave: save <name>
func (h *Handler) handleSave(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: save <name> (e.g., 'save my_pattern')")
	}

	// Join remaining parts as the name (allows spaces)
	name := strings.Join(parts[1:], " ")

	// Check if pattern file already exists (warn about overwrite)
	// Note: We replicate sanitization logic here since it's not exported
	sanitized := strings.ReplaceAll(name, " ", "_")
	filename := sanitized + ".json"
	patternPath := filepath.Join(sequence.PatternsDir, filename)
	if _, err := os.Stat(patternPath); err == nil {
		fmt.Printf("⚠️  Warning: Pattern '%s' already exists and will be overwritten.\n", name)
	}

	// Warn if global CC values exist (they won't be saved)
	globalCC := h.pattern.GetAllGlobalCC()
	if len(globalCC) > 0 {
		fmt.Println("⚠️  Warning: Global CC values will not be saved (they are transient).")
		fmt.Print("   Affected CC numbers: ")
		first := true
		for ccNum := range globalCC {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("CC#%d", ccNum)
			first = false
		}
		fmt.Println()
		fmt.Println("   Use 'cc-apply <cc-number>' to convert global CC to per-step automation before saving.")
		fmt.Println()
	}

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

	fmt.Printf("Loaded pattern '%s' (Tempo: %d BPM, Length: %d steps)\n", name, loadedPattern.BPM, loadedPattern.Length())
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

	// Warn about destructive operation
	fmt.Printf("⚠️  Warning: This will permanently delete pattern '%s'.\n", name)

	err := sequence.Delete(name)
	if err != nil {
		return fmt.Errorf("failed to delete pattern: %w", err)
	}

	fmt.Printf("Deleted pattern '%s'\n", name)
	return nil
}

// handleAI: ai [prompt] - execute AI prompt inline or enter interactive session
func (h *Handler) handleAI(parts []string) error {
	// Check if AI client is available
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	// Two modes:
	// 1. "ai" (no args) - enter interactive session
	// 2. "ai <prompt>" (with args) - execute inline (for batch scripts)

	if len(parts) == 1 {
		// Mode 1: Interactive session
		return h.handleAIInteractive()
	}

	// Mode 2: Inline execution
	// Join remaining parts as the prompt
	prompt := strings.Join(parts[1:], " ")
	return h.handleAIInline(prompt)
}

// handleAIInteractive enters an interactive AI session with readline
func (h *Handler) handleAIInteractive() error {
	// Clear any previous conversation history to start fresh
	h.aiClient.ClearHistory()

	fmt.Println("Entering AI session. Commands work directly. Type 'exit' to return to command mode.")
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

		// Empty line: show pattern
		if input == "" {
			fmt.Println(h.pattern.String())
			continue
		}

		// Check if input is a known command - if so, execute it directly without AI
		if h.isKnownCommand(input) {
			if err := h.ProcessCommand(input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}

		// Not a known command - send to AI
		if err := h.executeAIRequest(ctx, input); err != nil {
			fmt.Printf("AI error: %v\n", err)
		}

		fmt.Println()
	}
}

// handleAIInline executes a single AI prompt inline (for batch mode)
func (h *Handler) handleAIInline(prompt string) error {
	ctx := context.Background()
	return h.executeAIRequest(ctx, prompt)
}

// executeAIRequest sends a prompt to AI and executes the response
func (h *Handler) executeAIRequest(ctx context.Context, prompt string) error {
	// Send the entire pattern object to the AI session
	response, err := h.aiClient.Session(ctx, prompt, h.pattern)
	if err != nil {
		return err
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

	return nil
}

// isKnownCommand checks if the input starts with a known command
func (h *Handler) isKnownCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])
	knownCommands := []string{
		"set", "rest", "clear", "reset", "tempo", "velocity", "gate", "length",
		"humanize", "swing",
		"cc", "cc-step", "cc-clear", "cc-apply", "cc-show",
		"show", "verbose", "save", "load", "list", "delete",
		"clear-chat", "help", "quit",
	}

	for _, known := range knownCommands {
		if cmd == known {
			return true
		}
	}

	return false
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

	patternLen := h.pattern.Length()

	helpText := fmt.Sprintf(`Available commands:

PATTERN EDITING:
  set <step> <note|rest> [vel:<val>] [gate:<%%>] [dur:<steps>]
                          Set a step to play a note or rest
  rest <step>             Set a step to rest/silence
  velocity <step> <val>   Set step velocity 0-127
  gate <step> <percent>   Set step gate length 1-100%%
  humanize <type> <amt>   Add random variation (velocity/timing/gate)
  swing <percent>         Add swing/groove (0-75%%)
  length <steps>          Set pattern length
  clear                   Clear all steps to rests
  reset                   Reset to default pattern
  tempo <bpm>             Change tempo

CC AUTOMATION:
  cc <cc-num> <val>       Set global CC value (transient)
  cc-step <step> <cc> <val>  Set per-step CC automation
  cc-clear <step> [cc]    Clear CC automation from a step
  cc-apply <cc-num>       Apply global CC to all steps with notes
  cc-show                 Display all CC automation

PATTERN STORAGE:
  show                    Display current pattern
  save <name>             Save current pattern
  load <name>             Load a saved pattern
  list                    List all saved patterns
  delete <name>           Delete a saved pattern

AI FEATURES (AI: %s):
  ai [prompt]             Enter AI session or execute inline prompt
  clear-chat              Clear AI conversation history
  models                  List available AI models
  model <id>              Switch AI model (haiku/sonnet/opus)

MODEL COMPARISON:
  compare <prompt>        Run prompt against all models, save results
  compare-list            List saved comparisons
  compare-view <id>       View comparison details
  compare-load <id> <model>  Load pattern from comparison
  compare-delete <id>     Delete a comparison
  compare-rate <id> <model> <criteria> <score>
                          Rate a model (criteria: rhythmic/dynamics/genre/overall/all)
  blind <id>              Enter blind evaluation mode

OTHER:
  verbose [on|off]        Toggle verbose step output
  help                    Show this help message
  quit                    Exit the program

Notes: Steps 1-%d | Velocity 0-127 | Gate 1-100%% | CC 0-127
Patterns in 'patterns/', comparisons in 'comparisons/'`, aiStatus, patternLen)

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

// handleCompare: compare <prompt> - Run comparison test against all models
func (h *Handler) handleCompare(parts []string) error {
	// Check if AI client is available
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	if len(parts) < 2 {
		return fmt.Errorf("usage: compare <prompt> (e.g., 'compare create a funky bass line')")
	}

	// Join remaining parts as the prompt
	prompt := strings.Join(parts[1:], " ")

	// Validate prompt is not empty
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	fmt.Printf("Running comparison test with prompt: %q\n", prompt)
	fmt.Printf("Testing %d models: %s\n\n", len(comparison.AvailableModels), strings.Join(comparison.GetModelIDs(), ", "))

	ctx := context.Background()

	// Progress callback
	progress := func(modelName string, status string) {
		switch status {
		case "running":
			fmt.Printf("  %s: generating...\n", modelName)
		case "success":
			fmt.Printf("  %s: done\n", modelName)
		case "error":
			fmt.Printf("  %s: error\n", modelName)
		case "parse_error":
			fmt.Printf("  %s: parse error\n", modelName)
		}
	}

	// Run comparison
	comp, err := comparison.RunComparison(ctx, prompt, h.aiClient, progress)
	if err != nil {
		return fmt.Errorf("comparison failed: %w", err)
	}

	// Save results
	if err := comparison.SaveComparison(comp); err != nil {
		return fmt.Errorf("failed to save comparison: %w", err)
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Comparison complete (ID: %s)\n", comp.ID)
	fmt.Printf("Status: %s\n", comp.Status)
	fmt.Println()

	// Print results summary
	fmt.Println("Results:")
	for _, result := range comp.Results {
		statusSymbol := "✓"
		if result.Status != comparison.ResultSuccess {
			statusSymbol = "✗"
		}
		fmt.Printf("  %s %s: %s (%dms)\n", statusSymbol, result.ModelDisplayName, result.Status, result.DurationMs)
		if result.Status == comparison.ResultSuccess && result.Pattern != nil {
			fmt.Printf("      Commands: %d, Steps with notes: %d\n", len(result.Commands), len(result.Pattern.Steps))
		}
		if result.Error != "" {
			fmt.Printf("      Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\nSaved to: %s/%s.json\n", comparison.ComparisonsDir, comp.ID)
	fmt.Println("Use 'compare-view <id>' to see full details")

	return nil
}

// handleCompareList: compare-list - List all saved comparisons
func (h *Handler) handleCompareList(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: compare-list")
	}

	ids, err := comparison.ListComparisons()
	if err != nil {
		return fmt.Errorf("failed to list comparisons: %w", err)
	}

	if len(ids) == 0 {
		fmt.Println("No saved comparisons found")
		fmt.Println("Use 'compare <prompt>' to run a comparison test")
		return nil
	}

	fmt.Printf("Saved comparisons (%d):\n", len(ids))
	for _, id := range ids {
		// Load each comparison to show summary
		comp, err := comparison.LoadComparison(id)
		if err != nil {
			fmt.Printf("  %s (error loading)\n", id)
			continue
		}

		// Truncate prompt for display
		promptPreview := comp.Prompt
		if len(promptPreview) > 40 {
			promptPreview = promptPreview[:37] + "..."
		}

		successCount := len(comp.SuccessfulResults())
		fmt.Printf("  %s [%s] %d/%d models - %q\n", id, comp.Status, successCount, len(comp.Results), promptPreview)
	}

	return nil
}

// handleCompareView: compare-view <id> - View comparison details
func (h *Handler) handleCompareView(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: compare-view <id> (e.g., 'compare-view 20241206-143022')")
	}

	id := parts[1]
	comp, err := comparison.LoadComparison(id)
	if err != nil {
		return err
	}

	// Header
	fmt.Printf("Comparison: %s\n", comp.ID)
	fmt.Printf("Created: %s\n", comp.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Status: %s\n", comp.Status)
	fmt.Printf("Prompt: %q\n\n", comp.Prompt)

	// Results
	fmt.Println("Results:")
	for _, result := range comp.Results {
		statusSymbol := "✓"
		if result.Status != comparison.ResultSuccess {
			statusSymbol = "✗"
		}

		fmt.Printf("\n  %s %s (%s, %dms)\n", statusSymbol, result.ModelDisplayName, result.Status, result.DurationMs)

		if result.Error != "" {
			fmt.Printf("    Error: %s\n", result.Error)
		}

		if result.Status == comparison.ResultSuccess {
			fmt.Printf("    Commands (%d):\n", len(result.Commands))
			for _, cmd := range result.Commands {
				fmt.Printf("      %s\n", cmd)
			}

			if result.Pattern != nil {
				fmt.Printf("    Pattern: %d steps, %d notes\n", result.Pattern.Length, len(result.Pattern.Steps))
			}
		}

		if result.RawResponse != "" {
			fmt.Printf("    Raw response (truncated):\n      %s\n", truncateString(result.RawResponse, 200))
		}

		// Show rating if exists
		if comp.HasRating(result.Model) {
			rating := comp.Ratings[result.Model]
			fmt.Printf("    Rating: rhythmic=%d, dynamics=%d, genre=%d, overall=%d\n",
				rating.RhythmicInterest, rating.VelocityDynamics, rating.GenreAccuracy, rating.Overall)
		}
	}

	fmt.Printf("\nUse 'compare-load %s <model>' to load a pattern (models: %s)\n", id, strings.Join(comparison.GetModelIDs(), ", "))

	return nil
}

// handleCompareLoad: compare-load <id> <model> - Load a model's pattern from comparison
func (h *Handler) handleCompareLoad(parts []string) error {
	if len(parts) != 3 {
		return fmt.Errorf("usage: compare-load <id> <model> (e.g., 'compare-load 20241206-143022 haiku')")
	}

	id := parts[1]
	modelID := strings.ToLower(parts[2])

	// Load comparison
	comp, err := comparison.LoadComparison(id)
	if err != nil {
		return err
	}

	// Find model config to get API model ID
	modelConfig, found := comparison.GetModelByID(modelID)
	if !found {
		return fmt.Errorf("unknown model: %s (available: %s)", modelID, strings.Join(comparison.GetModelIDs(), ", "))
	}

	// Find result for this model
	result := comp.GetResultByModelID(string(modelConfig.APIModel))
	if result == nil {
		return fmt.Errorf("model '%s' not found in comparison %s", modelID, id)
	}

	if result.Status != comparison.ResultSuccess {
		return fmt.Errorf("model '%s' failed in this comparison (status: %s)", modelID, result.Status)
	}

	if result.Pattern == nil {
		return fmt.Errorf("no pattern data for model '%s' in comparison %s", modelID, id)
	}

	// Convert PatternFile to Pattern and copy to current pattern
	loadedPattern, err := sequence.FromPatternFile(result.Pattern)
	if err != nil {
		return fmt.Errorf("failed to load pattern: %w", err)
	}

	h.pattern.CopyFrom(loadedPattern)

	fmt.Printf("Loaded %s's pattern from comparison %s\n", modelConfig.DisplayName, id)
	fmt.Printf("Pattern: %d steps, tempo %d BPM\n", loadedPattern.Length(), loadedPattern.BPM)

	return nil
}

// handleCompareDelete: compare-delete <id> - Delete a saved comparison
func (h *Handler) handleCompareDelete(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: compare-delete <id> (e.g., 'compare-delete 20241206-143022')")
	}

	id := parts[1]

	// Verify it exists first
	_, err := comparison.LoadComparison(id)
	if err != nil {
		return err
	}

	if err := comparison.DeleteComparison(id); err != nil {
		return err
	}

	fmt.Printf("Deleted comparison %s\n", id)
	return nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// handleModels: models - List available AI models
func (h *Handler) handleModels(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: models")
	}

	fmt.Println("Available AI models:")
	for _, m := range comparison.AvailableModels {
		activeMarker := ""
		if m.ID == h.currentModel {
			activeMarker = " [active]"
		}
		fmt.Printf("  %s - %s (%s)%s\n", m.ID, m.DisplayName, m.Provider, activeMarker)
	}

	fmt.Println("\nUse 'model <id>' to switch models")
	return nil
}

// handleModel: model <id> - Switch AI model
func (h *Handler) handleModel(parts []string) error {
	if h.aiClient == nil {
		return fmt.Errorf("AI not available. Set ANTHROPIC_API_KEY environment variable to enable AI features")
	}

	if len(parts) != 2 {
		return fmt.Errorf("usage: model <id> (e.g., 'model sonnet')\nAvailable: %s", strings.Join(comparison.GetModelIDs(), ", "))
	}

	modelID := strings.ToLower(parts[1])

	// Validate model ID
	modelConfig, found := comparison.GetModelByID(modelID)
	if !found {
		return fmt.Errorf("unknown model: %s (available: %s)", modelID, strings.Join(comparison.GetModelIDs(), ", "))
	}

	// Check if already using this model
	if modelID == h.currentModel {
		fmt.Printf("Already using %s\n", modelConfig.DisplayName)
		return nil
	}

	// Switch model
	h.aiClient.SetModel(modelConfig.APIModel)
	h.currentModel = modelID

	fmt.Printf("Switched to %s\n", modelConfig.DisplayName)
	return nil
}

// handleBlind: blind <id> - Enter blind evaluation mode for a comparison
func (h *Handler) handleBlind(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: blind <id> (e.g., 'blind 20241206-143022')")
	}

	id := parts[1]

	// Load comparison
	comp, err := comparison.LoadComparison(id)
	if err != nil {
		return err
	}

	// Get successful results only
	successfulResults := comp.SuccessfulResults()
	if len(successfulResults) == 0 {
		return fmt.Errorf("no successful results in comparison %s", id)
	}

	if len(successfulResults) == 1 {
		fmt.Println("Warning: Only one model succeeded in this comparison. Blind evaluation is less meaningful.")
	}

	// Collect model IDs from successful results
	modelIDs := make([]string, len(successfulResults))
	for i, r := range successfulResults {
		modelIDs[i] = r.Model
	}

	// Create blind session with randomized labels
	session := comparison.NewBlindSession(id, modelIDs)

	fmt.Printf("Entering blind evaluation mode for comparison %s\n", id)
	fmt.Printf("Prompt: %q\n\n", comp.Prompt)
	fmt.Printf("Patterns to evaluate (%d):\n", len(session.Labels))
	for _, label := range session.Labels {
		fmt.Printf("  Pattern %s\n", label)
	}
	fmt.Println()
	fmt.Println("Commands in blind mode:")
	fmt.Println("  load <label>     - Load pattern into active playback (e.g., 'load A')")
	fmt.Println("  rate <label> <1-5> - Rate a pattern (e.g., 'rate A 4')")
	fmt.Println("  status           - Show rating progress")
	fmt.Println("  reveal           - Reveal which model made which pattern (after all rated)")
	fmt.Println("  exit             - Exit blind mode without saving")
	fmt.Println()

	// Enter blind mode loop
	return h.blindModeLoop(session, comp)
}

// blindModeLoop runs the interactive blind evaluation mode
func (h *Handler) blindModeLoop(session *comparison.BlindSession, comp *comparison.Comparison) error {
	rl, err := readline.New("blind> ")
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Println("\nExiting blind mode.")
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit":
			fmt.Println("Exiting blind mode (ratings not saved)")
			return nil

		case "load":
			if len(parts) != 2 {
				fmt.Println("Usage: load <label> (e.g., 'load A')")
				continue
			}
			label := strings.ToUpper(parts[1])
			if err := h.blindLoadPattern(session, comp, label); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "rate":
			if len(parts) != 3 {
				fmt.Println("Usage: rate <label> <1-5> (e.g., 'rate A 4')")
				continue
			}
			label := strings.ToUpper(parts[1])
			score, err := strconv.Atoi(parts[2])
			if err != nil || score < 1 || score > 5 {
				fmt.Println("Score must be 1-5")
				continue
			}
			if err := h.blindRatePattern(session, label, score); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "status":
			h.blindShowStatus(session)

		case "reveal":
			if !session.IsComplete() {
				fmt.Printf("Cannot reveal yet - %d/%d patterns rated\n", session.RatedCount(), session.TotalCount())
				fmt.Println("Rate all patterns first, then reveal")
				continue
			}
			h.blindReveal(session, comp)
			return nil

		case "show":
			fmt.Println(h.pattern.String())

		default:
			fmt.Printf("Unknown blind mode command: %s\n", cmd)
			fmt.Println("Commands: load, rate, status, reveal, show, exit")
		}
	}
}

// blindLoadPattern loads a pattern by label into the active pattern
func (h *Handler) blindLoadPattern(session *comparison.BlindSession, comp *comparison.Comparison, label string) error {
	modelID, exists := session.GetModelIDByLabel(label)
	if !exists {
		return fmt.Errorf("unknown label: %s (available: %s)", label, strings.Join(session.Labels, ", "))
	}

	result := comp.GetResultByModelID(modelID)
	if result == nil || result.Pattern == nil {
		return fmt.Errorf("no pattern data for label %s", label)
	}

	loadedPattern, err := sequence.FromPatternFile(result.Pattern)
	if err != nil {
		return fmt.Errorf("failed to load pattern: %w", err)
	}

	h.pattern.CopyFrom(loadedPattern)

	ratingStatus := ""
	if session.IsRated(label) {
		ratingStatus = fmt.Sprintf(" (rated: %d)", session.GetRating(label))
	}
	fmt.Printf("Loaded Pattern %s%s\n", label, ratingStatus)

	return nil
}

// blindRatePattern rates a pattern by label
func (h *Handler) blindRatePattern(session *comparison.BlindSession, label string, score int) error {
	if !session.RateLabel(label, score) {
		return fmt.Errorf("unknown label: %s (available: %s)", label, strings.Join(session.Labels, ", "))
	}

	fmt.Printf("Rated Pattern %s: %d/5\n", label, score)
	fmt.Printf("Progress: %d/%d patterns rated\n", session.RatedCount(), session.TotalCount())

	if session.IsComplete() {
		fmt.Println("\nAll patterns rated! Use 'reveal' to see which model made each pattern.")
	}

	return nil
}

// blindShowStatus shows the current rating status
func (h *Handler) blindShowStatus(session *comparison.BlindSession) {
	fmt.Printf("Rating progress: %d/%d\n", session.RatedCount(), session.TotalCount())
	for _, label := range session.Labels {
		if session.IsRated(label) {
			fmt.Printf("  Pattern %s: %d/5\n", label, session.GetRating(label))
		} else {
			fmt.Printf("  Pattern %s: not rated\n", label)
		}
	}
}

// blindReveal shows the label-to-model mapping and saves ratings
func (h *Handler) blindReveal(session *comparison.BlindSession, comp *comparison.Comparison) {
	fmt.Println("\n=== REVEAL ===")

	results := session.GetRevealResults(comp)

	// Sort by rating (highest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Rating > results[i].Rating {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	fmt.Println("Results (sorted by rating):")
	for _, r := range results {
		fmt.Printf("  Pattern %s = %s (rating: %d/5)\n", r.Label, r.DisplayName, r.Rating)
	}

	// Save ratings to comparison
	if comp.Ratings == nil {
		comp.Ratings = make(map[string]*comparison.Rating)
	}

	for _, label := range session.Labels {
		modelID, _ := session.GetModelIDByLabel(label)
		score := session.GetRating(label)

		// Create rating with overall score from blind evaluation
		comp.Ratings[modelID] = &comparison.Rating{
			Overall: score,
		}
	}

	// Save updated comparison
	if err := comparison.SaveComparison(comp); err != nil {
		fmt.Printf("\nWarning: Failed to save ratings: %v\n", err)
	} else {
		fmt.Printf("\nRatings saved to comparison %s\n", comp.ID)
	}

	fmt.Println("\nExiting blind mode.")
}

// handleCompareRate: compare-rate <id> <model> <criteria> <score> - Rate a model's output
func (h *Handler) handleCompareRate(parts []string) error {
	if len(parts) != 5 {
		return fmt.Errorf("usage: compare-rate <id> <model> <criteria> <score>\n" +
			"  criteria: rhythmic, dynamics, genre, overall, all\n" +
			"  score: 1-5\n" +
			"  e.g., 'compare-rate 20241206-143022 haiku overall 4'")
	}

	id := parts[1]
	modelID := strings.ToLower(parts[2])
	criteria := strings.ToLower(parts[3])
	scoreStr := parts[4]

	// Parse score
	score, err := strconv.Atoi(scoreStr)
	if err != nil || !comparison.IsValidScore(score) {
		return fmt.Errorf("score must be 1-5, got: %s", scoreStr)
	}

	// Validate criteria
	if !comparison.IsValidCriteria(criteria) {
		return fmt.Errorf("invalid criteria: %s (valid: %s)", criteria, strings.Join(comparison.ValidRatingCriteria, ", "))
	}

	// Load comparison
	comp, err := comparison.LoadComparison(id)
	if err != nil {
		return err
	}

	// Find model config
	modelConfig, found := comparison.GetModelByID(modelID)
	if !found {
		return fmt.Errorf("unknown model: %s (available: %s)", modelID, strings.Join(comparison.GetModelIDs(), ", "))
	}

	// Check if model is in comparison
	result := comp.GetResultByModelID(string(modelConfig.APIModel))
	if result == nil {
		return fmt.Errorf("model '%s' not found in comparison %s", modelID, id)
	}

	// Initialize ratings map if needed
	if comp.Ratings == nil {
		comp.Ratings = make(map[string]*comparison.Rating)
	}

	// Get or create rating for this model
	rating := comp.Ratings[result.Model]
	if rating == nil {
		rating = comparison.NewRating()
		comp.Ratings[result.Model] = rating
	}

	// Set the criteria
	if !rating.SetCriteria(criteria, score) {
		return fmt.Errorf("failed to set criteria: %s", criteria)
	}

	// Save comparison
	if err := comparison.SaveComparison(comp); err != nil {
		return fmt.Errorf("failed to save rating: %w", err)
	}

	if criteria == "all" {
		fmt.Printf("Rated %s in comparison %s: all criteria = %d\n", modelConfig.DisplayName, id, score)
	} else {
		fmt.Printf("Rated %s in comparison %s: %s = %d\n", modelConfig.DisplayName, id, criteria, score)
	}

	// Show full rating if complete
	if rating.IsComplete() {
		fmt.Printf("  Full rating: rhythmic=%d, dynamics=%d, genre=%d, overall=%d\n",
			rating.RhythmicInterest, rating.VelocityDynamics, rating.GenreAccuracy, rating.Overall)
	}

	return nil
}
