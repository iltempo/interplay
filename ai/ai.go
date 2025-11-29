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

// Client wraps the Claude API client
type Client struct {
	client anthropic.Client
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
func (c *Client) Chat(ctx context.Context, question string, currentPattern string) (string, error) {
	userMessage := fmt.Sprintf("Current pattern:\n%s\n\nQuestion: %s", currentPattern, question)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: chatSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
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

	return strings.TrimSpace(responseText), nil
}
