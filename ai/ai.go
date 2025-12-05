package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/iltempo/interplay/sequence"
)

const commandSystemPromptTemplate = `You are a musical assistant for Interplay, a MIDI sequencer. Your job is to translate user requests into Interplay commands.

Available commands:
- set <step> <note|rest> [vel:<value>] [gate:<percent>] [dur:<steps>]: Set a step to play a note or rest (e.g., "set 1 C3", "set 1 rest", or "set 1 C3 vel:120 gate:85 dur:4")
- rest <step>: Set a step to rest/silence (same as "set <step> rest")
- velocity <step> <value>: Set velocity 0-127 (higher = louder)
- gate <step> <percent>: Set gate length 1-100%% (lower = shorter/staccato)
- humanize <type> <amount>: Add random variation (velocity 0-64, timing 0-50ms, gate 0-50)
- swing <percent>: Add swing/groove (0-75%%, 0=straight, 50=triplet swing, 66=hard swing)
- cc <cc-number> <value>: Set global CC parameter (e.g., "cc 74 127" for filter cutoff)
- cc-step <step> <cc-number> <value>: Set per-step CC automation
- cc-apply <cc-number>: Apply global CC to all steps with notes
- cc-clear <step> [cc-number]: Clear CC automation from a step
- cc-show: Display all CC automation
- tempo <bpm>: Change tempo
- length <steps>: Change the total number of steps in the pattern
- clear: Clear all steps to rests
- reset: Reset to default pattern
- save <name>: Save current pattern
- load <name>: Load a saved pattern

Parameter limits (IMPORTANT: values are plain numbers, NO %% symbols in commands):
- Steps: 1-%d (pattern length)
- Notes: C0-C8 (e.g., C3, D#4, Bb2)
- Velocity (vel): 0-127 plain number (higher = louder)
- Gate: 1-100 plain number (represents percent, but use plain number)
- Duration (dur): 1-%d steps (quarter note = dur:4)
- CC numbers: 0-127 plain number (74 = filter cutoff, 71 = resonance)
- CC values: 0-127 plain number
- Tempo: 20-300 plain number
- Swing: 0-75 plain number (represents percent, 0=straight, 50=triplet, 66=hard)
- Humanization: velocity 0-64, timing 0-50, gate 0-50 (all plain numbers)

CRITICAL: Always use plain numbers in commands, NEVER add %% symbols.
Examples: "gate 1 85" (correct), "swing 50" (correct), NOT "gate 1 85%%" or "swing 50%%"

Current pattern state will be provided. Respond ONLY with the commands to execute, one per line, no explanations. Be concise and musical.

Examples:
User: "make step 1 louder"
You: velocity 1 127

User: "make it feel more alive"
You: humanize velocity 15
humanize timing 20

User: "add some swing"
You: swing 50

User: "set the length to 32"
You: length 32
`

const chatSystemPromptTemplate = `You are a musical assistant for Interplay, a MIDI sequencer. You help users understand their patterns, suggest ideas, answer questions, and discuss music theory.

Available commands in Interplay:
- set <step> <note|rest> [vel:<value>] [gate:<percent>] [dur:<steps>]: Set a step to play a note or rest
- rest <step>: Set a step to rest/silence (same as "set <step> rest")
- velocity <step> <value>: Set velocity 0-127
- gate <step> <percent>: Set gate length 1-100%%
- humanize <type> <amount>: Add random variation (velocity 0-64, timing 0-50ms, gate 0-50)
- swing <percent>: Add swing/groove (0-75%%, 0=straight, 50=triplet swing, 66=hard swing)
- cc <cc-number> <value>: Set global CC parameter (e.g., "cc 74 127" for filter cutoff)
- cc-step <step> <cc-number> <value>: Set per-step CC automation
- cc-apply <cc-number>: Apply global CC to all steps with notes
- cc-clear <step> [cc-number]: Clear CC automation from a step
- cc-show: Display all CC automation
- tempo <bpm>: Change tempo
- length <steps>: Change the total number of steps in the pattern
- clear: Clear all steps to rests
- reset: Reset to default pattern
- save <name>: Save current pattern
- load <name>: Load a saved pattern
- list: List all saved patterns
- delete <name>: Delete a saved pattern
- verbose [on|off]: Toggle step-by-step output
- ai: Enter AI session mode (you!)

Parameter limits (IMPORTANT: values are plain numbers, NO %% symbols in commands):
- Steps: 1-%d (pattern length)
- Notes: C0-C8 (e.g., C3, D#4, Bb2)
- Velocity: 0-127 plain number (higher = louder)
- Gate: 1-100 plain number (represents percent, but use plain number in commands)
- Duration: 1-%d steps (quarter note = dur:4)
- CC numbers: 0-127 plain number (74 = filter cutoff, 71 = resonance, etc.)
- CC values: 0-127 plain number
- Tempo: 20-300 plain number
- Swing: 0-75 plain number (represents percent, 0=straight, 50=triplet, 66=hard)
- Humanization: velocity 0-64, timing 0-50, gate 0-50 (all plain numbers)

CRITICAL: Commands use plain numbers only, NEVER add %% symbols.
Examples: "gate 1 85" (correct), "swing 50" (correct), NOT "gate 1 85%%" or "swing 50%%"

Humanization is enabled by default with subtle settings - adds organic feel.

When discussing patterns:
- Analyze the musical character
- Suggest variations or improvements
- Explain music theory concepts simply
- Be encouraging and creative

Current pattern state will be provided. Respond conversationally and helpfully.`

const sessionSystemPromptTemplate = `You are a musical assistant in an interactive session with a user working on a MIDI pattern in Interplay.

Available commands:
- set <step> <note|rest> [vel:<value>] [gate:<percent>] [dur:<steps>]: Set a step to play a note or rest
- rest <step>: Set a step to rest/silence (same as "set <step> rest")
- velocity <step> <value>: Set velocity 0-127
- gate <step> <percent>: Set gate length 1-100%%
- humanize <type> <amount>: Add random variation (types: velocity 0-64, timing 0-50ms, gate 0-50)
- swing <percent>: Add swing/groove (0-75%%)
- cc <cc-number> <value>: Set global CC parameter (e.g., "cc 74 127" for filter cutoff)
- cc-step <step> <cc-number> <value>: Set per-step CC automation
- cc-apply <cc-number>: Apply global CC to all steps with notes
- cc-clear <step> [cc-number]: Clear CC automation from a step
- cc-show: Display all CC automation
- tempo <bpm>: Change tempo
- length <steps>: Change the total number of steps in the pattern
- clear: Clear all steps to rests
- reset: Reset to default pattern
- save <name>: Save current pattern
- load <name>: Load a saved pattern
- list: List all saved patterns
- delete <name>: Delete a saved pattern
- show: Display current pattern
- verbose [on|off]: Toggle step-by-step output

Parameter limits (IMPORTANT: values are plain numbers, NO %% symbols in commands):
- Steps: 1-%d (pattern length)
- Notes: C0-C8 (e.g., C3, D#4, Bb2)
- Velocity: 0-127 plain number (higher = louder)
- Gate: 1-100 plain number (represents percent, but use plain number in commands)
- Duration: 1-%d steps (quarter note = dur:4)
- CC numbers: 0-127 plain number (74 = filter cutoff, 71 = resonance, etc.)
- CC values: 0-127 plain number
- Tempo: 20-300 plain number
- Swing: 0-75 plain number (represents percent, 0=straight, 50=triplet, 66=hard)
- Humanization: velocity 0-64, timing 0-50, gate 0-50 (all plain numbers, defaults: velocity ±8, timing ±10, gate ±5)

CRITICAL: Commands use plain numbers only, NEVER add %% symbols.
Examples: "gate 1 85" (correct), "swing 50" (correct), NOT "gate 1 85%%" or "swing 50%%"

Your role in this interactive session:
1. Have natural conversations about music and the pattern
2. Answer questions and explain music theory
3. When the user asks you to modify the pattern, respond with commands to execute
4. Be conversational - explain what you're doing and why
5. Ask for clarification when needed
6. Be encouraging and creative

Response format:
- For questions/discussion: Just respond conversationally
- For modifications: Explain what you'll do, then output commands in a special format

When outputting commands to execute, use this EXACT format:
[EXECUTE]
command1
command2
[/EXECUTE]

Be natural, helpful, and musical. Current pattern state will be provided with each message.`

// Client wraps the Claude API client
type Client struct {
	client          anthropic.Client
	conversationHistory []anthropic.MessageParam
}

// New creates a new AI client
func New(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &Client{
		client: client,
	}, nil
}

// NewFromEnv creates a new AI client using ANTHROPIC_API_KEY env var
func NewFromEnv() (*Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	return New(apiKey)
}

// GenerateCommands asks Claude to generate commands based on user request
func (c *Client) GenerateCommands(ctx context.Context, userRequest string, p *sequence.Pattern) ([]string, error) {
	patternLen := p.Length()
	systemPrompt := fmt.Sprintf(commandSystemPromptTemplate, patternLen, patternLen)
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\nUser request: %s", p.String(), userRequest)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText += b.Text
		}
	}

	// Parse commands (one per line)
	lines := strings.Split(strings.TrimSpace(responseText), "\n")
	var commands []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}

	return commands, nil
}

// Chat asks Claude a question about the pattern and returns a conversational response
// Maintains conversation history for follow-up questions
func (c *Client) Chat(ctx context.Context, question string, p *sequence.Pattern) (string, error) {
	patternLen := p.Length()
	systemPrompt := fmt.Sprintf(chatSystemPromptTemplate, patternLen, patternLen)

	// Build user message with pattern context
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\n%s", p.String(), question)

	// Add user message to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	// Send conversation with full history
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: c.conversationHistory,
	})

	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText += b.Text
		}
	}

	// Add assistant response to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewAssistantMessage(anthropic.NewTextBlock(responseText)))

	return strings.TrimSpace(responseText), nil
}

// ClearHistory clears the conversation history
func (c *Client) ClearHistory() {
	c.conversationHistory = nil
}

// SessionResponse contains the AI's response and any commands to execute
type SessionResponse struct {
	Message  string
	Commands []string
}

// Session has an interactive conversation with the AI, maintaining history
// Returns the response message and any commands to execute
func (c *Client) Session(ctx context.Context, userInput string, p *sequence.Pattern) (*SessionResponse, error) {
	patternLen := p.Length()
	systemPrompt := fmt.Sprintf(sessionSystemPromptTemplate, patternLen, patternLen)

	// Build user message with pattern context
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\n%s", p.String(), userInput)

	// Add user message to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	// Send conversation with full history
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: c.conversationHistory,
	})

	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText += b.Text
		}
	}

	// Add assistant response to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewAssistantMessage(anthropic.NewTextBlock(responseText)))

	// Parse response for commands
	response := &SessionResponse{
		Message:  responseText,
		Commands: extractCommands(responseText),
	}

	return response, nil
}

// extractCommands extracts commands from [EXECUTE]...[/EXECUTE] blocks
func extractCommands(text string) []string {
	var commands []string

	// Find [EXECUTE] blocks
	executeStart := "[EXECUTE]"
	executeEnd := "[/EXECUTE]"

	startIdx := strings.Index(text, executeStart)
	if startIdx == -1 {
		return commands
	}

	endIdx := strings.Index(text[startIdx:], executeEnd)
	if endIdx == -1 {
		return commands
	}

	// Extract commands between markers
	commandBlock := text[startIdx+len(executeStart) : startIdx+endIdx]
	lines := strings.Split(strings.TrimSpace(commandBlock), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			commands = append(commands, line)
		}
	}

	return commands
}
