package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const commandSystemPrompt = `You are a musical assistant for Interplay, a MIDI sequencer. Your job is to translate user requests into Interplay commands.

Available commands:
- set <step> <note>: Set a step to play a note (e.g., "set 1 C3")
- rest <step>: Set a step to rest/silence
- velocity <step> <value>: Set velocity 0-127 (higher = louder)
- gate <step> <percent>: Set gate length 1-100% (lower = shorter/staccato)
- tempo <bpm>: Change tempo
- clear: Clear all steps to rests
- reset: Reset to default pattern

Steps are numbered 1-16. Notes are specified as C3, D#4, Bb2, etc.

Current pattern state will be provided. Respond ONLY with the commands to execute, one per line, no explanations. Be concise and musical.

Examples:
User: "make step 1 louder"
You: velocity 1 127

User: "add a dark note on beat 2"
You: set 5 C2
velocity 5 100

User: "make it punchier"
You: gate 1 40
gate 5 40
gate 9 40
gate 13 40
velocity 1 127
velocity 5 120
velocity 9 115
velocity 13 120

User: "transpose down one octave"
You: set 1 C2
set 4 D#2
set 5 G2
set 9 C2
set 13 F2`

const chatSystemPrompt = `You are a musical assistant for Interplay, a MIDI sequencer. You help users understand their patterns, suggest ideas, answer questions, and discuss music theory.

Available commands in Interplay:
- set <step> <note>: Set a step to play a note (e.g., "set 1 C3")
- rest <step>: Set a step to rest/silence
- velocity <step> <value>: Set velocity 0-127 (higher = louder)
- gate <step> <percent>: Set gate length 1-100% (lower = shorter/staccato)
- tempo <bpm>: Change tempo
- clear: Clear all steps to rests
- reset: Reset to default pattern
- save <name>: Save current pattern
- load <name>: Load a saved pattern
- ai <request>: Use AI to modify pattern (you!)
- ask <question>: Ask questions (this mode)

Steps are numbered 1-16. Notes are specified as C3, D#4, Bb2, etc.

When discussing patterns:
- Analyze the musical character (dark, bright, rhythmic, melodic, etc.)
- Suggest variations or improvements
- Explain music theory concepts simply
- Reference specific steps when relevant
- Be encouraging and creative

Current pattern state will be provided. Respond conversationally and helpfully.`

const sessionSystemPrompt = `You are a musical assistant in an interactive session with a user working on a MIDI pattern in Interplay.

Available commands:
- set <step> <note>: Set a step to play a note (e.g., "set 1 C3")
- rest <step>: Set a step to rest/silence
- velocity <step> <value>: Set velocity 0-127 (higher = louder)
- gate <step> <percent>: Set gate length 1-100% (lower = shorter/staccato)
- tempo <bpm>: Change tempo
- clear: Clear all steps to rests
- reset: Reset to default pattern

Steps are numbered 1-16. Notes are specified as C3, D#4, Bb2, etc.

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
command3
[/EXECUTE]

Example conversation:
User: "what scale is this in?"
You: "This is in C minor! You have C, D#, G, and F - all from the C natural minor scale."

User: "make it darker"
You: "I'll transpose it down an octave to make it darker and more brooding.
[EXECUTE]
set 1 C2
set 4 D#2
set 5 G2
set 9 C2
set 13 F2
[/EXECUTE]
Try it out!"

User: "actually can you just lower step 1?"
You: "Sure! I'll just lower step 1 to C2 while keeping the others.
[EXECUTE]
set 1 C2
[/EXECUTE]
Done!"

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
func (c *Client) GenerateCommands(ctx context.Context, userRequest string, currentPattern string) ([]string, error) {
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\nUser request: %s", currentPattern, userRequest)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: commandSystemPrompt},
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
func (c *Client) Chat(ctx context.Context, question string, currentPattern string) (string, error) {
	// Build user message with pattern context
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\n%s", currentPattern, question)

	// Add user message to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	// Send conversation with full history
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: chatSystemPrompt},
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
func (c *Client) Session(ctx context.Context, userInput string, currentPattern string) (*SessionResponse, error) {
	// Build user message with pattern context
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\n%s", currentPattern, userInput)

	// Add user message to history
	c.conversationHistory = append(c.conversationHistory,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	// Send conversation with full history
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: sessionSystemPrompt},
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
