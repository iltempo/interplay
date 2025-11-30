package ai

import (
	"reflect"
	"testing"
)

// TestExtractCommands tests the extraction of commands from [EXECUTE] blocks
func TestExtractCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Single command",
			input: `I'll set step 1 to C4.
[EXECUTE]
set 1 C4
[/EXECUTE]
Done!`,
			expected: []string{"set 1 C4"},
		},
		{
			name: "Multiple commands",
			input: `Let me transpose it down.
[EXECUTE]
set 1 C2
set 5 C2
set 9 G1
[/EXECUTE]
Try it out!`,
			expected: []string{"set 1 C2", "set 5 C2", "set 9 G1"},
		},
		{
			name: "Commands with extra whitespace",
			input: `[EXECUTE]
  set 1 C4

  tempo 120
  velocity 1 100
[/EXECUTE]`,
			expected: []string{"set 1 C4", "tempo 120", "velocity 1 100"},
		},
		{
			name:     "No execute block",
			input:    "This is just a conversational response with no commands.",
			expected: nil,
		},
		{
			name:     "Empty execute block",
			input:    "[EXECUTE]\n\n[/EXECUTE]",
			expected: nil,
		},
		{
			name:     "Unclosed execute block",
			input:    "[EXECUTE]\nset 1 C4\n",
			expected: nil,
		},
		{
			name: "Execute block with no opening tag",
			input: `set 1 C4
[/EXECUTE]`,
			expected: nil,
		},
		{
			name: "Complex response with commands",
			input: `I'll make it darker and more brooding by transposing down an octave.
[EXECUTE]
set 1 C2
set 4 D#2
set 5 G2
set 9 C2
set 13 F2
[/EXECUTE]
This should give it a much heavier feel!`,
			expected: []string{"set 1 C2", "set 4 D#2", "set 5 G2", "set 9 C2", "set 13 F2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCommands(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractCommands() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestClearHistory tests that conversation history is properly cleared
func TestClearHistory(t *testing.T) {
	// Create a client with a valid API key to initialize it
	client, err := New("sk-test-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Simulate adding some history (we can't actually add real messages without API calls)
	// but we can test that ClearHistory sets it to nil
	client.ClearHistory()

	if len(client.conversationHistory) != 0 {
		t.Errorf("After ClearHistory, length = %d, want 0", len(client.conversationHistory))
	}

	// Verify it's actually nil, not just empty
	if client.conversationHistory != nil {
		t.Error("After ClearHistory, conversationHistory should be nil")
	}
}

// TestNewFromEnv tests client creation from environment
func TestNewFromEnv(t *testing.T) {
	// Test with no API key set (should fail gracefully)
	t.Setenv("ANTHROPIC_API_KEY", "")

	client, err := NewFromEnv()
	if err == nil {
		t.Error("NewFromEnv() with empty API key should return error")
	}
	if client != nil {
		t.Error("NewFromEnv() with empty API key should return nil client")
	}
}

// TestNew tests client creation with API key
func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		wantError bool
	}{
		{
			name:      "Valid API key",
			apiKey:    "sk-ant-test-key-123",
			wantError: false,
		},
		{
			name:      "Empty API key",
			apiKey:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.apiKey)

			if tt.wantError {
				if err == nil {
					t.Error("New() should return error for empty API key")
				}
				if client != nil {
					t.Error("New() should return nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error: %v", err)
				}
				if client == nil {
					t.Error("New() should return non-nil client for valid API key")
				}
			}
		})
	}
}
