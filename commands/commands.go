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
	case "help":
		return h.handleHelp(parts)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// handleSet: set <step> <note> [vel:<value>] [gate:<percent>] [dur:<steps>]
func (h *Handler) handleSet(parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("usage: set <step> <note> [vel:<value>] [gate:<percent>] [dur:<steps>]\n" +
			"e.g., 'set 1 C4' or 'set 1 C4 vel:120 gate:85 dur:3'")
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

	fmt.Println("Reset to default 16-step pattern")
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
		// Send the entire pattern object to the AI session
		response, err := h.aiClient.Session(ctx, input, h.pattern)
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
  set <step> <note> [vel:<val>] [gate:<%%>] [dur:<steps>]
                          Set a step to play a note (e.g., 'set 1 C4')
                          Optional parameters can be combined in any order
                          (e.g., 'set 1 C4 vel:120 gate:85 dur:3')
  rest <step>             Set a step to rest/silence (e.g., 'rest 1')
  velocity <step> <val>   Set step velocity 0-127 (e.g., 'velocity 1 80')
  gate <step> <percent>   Set step gate length 1-100%% (e.g., 'gate 1 50')
  humanize <type> <amt>   Add random variation (e.g., 'humanize velocity 10')
                          Types: velocity (0-64), timing (0-50ms), gate (0-50)
                          Use 'humanize' alone to show current settings
  swing <percent>         Add swing/groove (e.g., 'swing 50' for triplet swing)
                          0 = straight, 50 = triplet, 66 = hard swing (0-75)
  cc <cc-num> <val>       Set global CC value (transient, not saved)
                          e.g., 'cc 74 127' sets filter cutoff to max
  cc-step <step> <cc> <val>
                          Set per-step CC automation (persistent, saved)
                          e.g., 'cc-step 1 74 127' sets filter on step 1
  cc-clear <step> [cc]    Clear CC automation from a step
                          e.g., 'cc-clear 1' clears all CC, 'cc-clear 1 74' clears CC#74
  cc-apply <cc-num>       Apply global CC to all steps with notes
                          e.g., 'cc-apply 74' converts global CC#74 to per-step
  cc-show                 Display all CC automation in table format
  length <steps>          Set pattern length (e.g., 'length 32')
  clear                   Clear all steps to rests
  reset                   Reset to default 16-step pattern
  tempo <bpm>             Change tempo (e.g., 'tempo 120')
  show                    Display current pattern (CC automation shown in brackets)
  verbose [on|off]        Toggle or set verbose step output
  save <name>             Save current pattern (e.g., 'save bass_line')
  load <name>             Load a saved pattern (e.g., 'load bass_line')
  list                    List all saved patterns
  delete <name>           Delete a saved pattern (e.g., 'delete bass_line')
  ai                      Enter interactive AI session (AI: %s)
                          All commands work directly in AI mode.
                          Natural language is sent to AI for pattern changes.
                          Type 'exit' to return to command mode.
  clear-chat              Clear AI conversation history
  help                    Show this help message
  quit                    Exit the program
  <enter>                 Show current pattern (same as 'show')

Notes: C4, D#5, Bb3, etc. | Steps: 1-%d | Duration: 1-%d steps (default 1)
Default velocity: 100 | Default gate: 90%% | CC numbers/values: 0-127
Patterns saved in 'patterns/' directory as JSON files.
AI features require ANTHROPIC_API_KEY environment variable.`, aiStatus, patternLen, patternLen)

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
