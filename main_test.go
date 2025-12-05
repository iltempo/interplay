package main

import (
	"strings"
	"testing"

	"github.com/iltempo/interplay/commands"
	"github.com/iltempo/interplay/sequence"
)

// mockVerboseController implements VerboseController for testing
type mockVerboseController struct {
	verbose bool
}

func (m *mockVerboseController) SetVerbose(v bool) { m.verbose = v }
func (m *mockVerboseController) IsVerbose() bool   { return m.verbose }

func TestProcessBatchInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSuccess bool
		wantExit    bool
	}{
		{
			name:        "empty input",
			input:       "",
			wantSuccess: true,
			wantExit:    false,
		},
		{
			name:        "comments only",
			input:       "# comment\n# another comment\n",
			wantSuccess: true,
			wantExit:    false,
		},
		{
			name:        "empty lines only",
			input:       "\n\n\n",
			wantSuccess: true,
			wantExit:    false,
		},
		{
			name:        "valid command",
			input:       "show\n",
			wantSuccess: true,
			wantExit:    false,
		},
		{
			name:        "exit command",
			input:       "exit\n",
			wantSuccess: true,
			wantExit:    true,
		},
		{
			name:        "quit command",
			input:       "quit\n",
			wantSuccess: true,
			wantExit:    true,
		},
		{
			name:        "mixed valid and comments",
			input:       "# Setup pattern\nshow\n# Done\n",
			wantSuccess: true,
			wantExit:    false,
		},
		{
			name:        "invalid command",
			input:       "invalid_command_xyz\n",
			wantSuccess: false,
			wantExit:    false,
		},
		{
			name:        "valid then invalid commands",
			input:       "show\ninvalid_command\n",
			wantSuccess: false,
			wantExit:    false,
		},
		{
			name:        "invalid then valid commands",
			input:       "invalid_command\nshow\n",
			wantSuccess: false,
			wantExit:    false,
		},
		{
			name:        "exit after error",
			input:       "invalid_command\nexit\n",
			wantSuccess: false,
			wantExit:    true,
		},
		{
			name:        "case insensitive exit",
			input:       "EXIT\n",
			wantSuccess: true,
			wantExit:    true,
		},
		{
			name:        "case insensitive quit",
			input:       "QUIT\n",
			wantSuccess: true,
			wantExit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			pattern := sequence.New(16)
			mockController := &mockVerboseController{}
			handler := commands.New(pattern, mockController)
			reader := strings.NewReader(tt.input)

			// Execute
			gotSuccess, gotExit := processBatchInput(reader, handler)

			// Verify
			if gotSuccess != tt.wantSuccess {
				t.Errorf("processBatchInput() success = %v, want %v", gotSuccess, tt.wantSuccess)
			}
			if gotExit != tt.wantExit {
				t.Errorf("processBatchInput() exit = %v, want %v", gotExit, tt.wantExit)
			}
		})
	}
}

func TestProcessBatchInput_CommandExecution(t *testing.T) {
	// Test that commands actually execute
	pattern := sequence.New(16)
	mockController := &mockVerboseController{}
	handler := commands.New(pattern, mockController)

	// Execute length command (easy to verify)
	input := "length 8\n"
	reader := strings.NewReader(input)
	success, exit := processBatchInput(reader, handler)

	if !success {
		t.Error("Expected length command to succeed")
	}
	if exit {
		t.Error("Expected no exit for length command")
	}

	// Verify length was actually set
	if pattern.Length() != 8 {
		t.Errorf("Expected length to be 8, got %d", pattern.Length())
	}
}

func TestProcessBatchInput_MultipleCommands(t *testing.T) {
	// Test multiple commands execute in sequence
	pattern := sequence.New(16)
	mockController := &mockVerboseController{}
	handler := commands.New(pattern, mockController)

	input := `# Set up pattern
length 8
clear
# Show result
show
`
	reader := strings.NewReader(input)
	success, exit := processBatchInput(reader, handler)

	if !success {
		t.Error("Expected all commands to succeed")
	}
	if exit {
		t.Error("Expected no exit")
	}

	// Verify commands were executed
	if pattern.Length() != 8 {
		t.Errorf("Expected length to be 8, got %d", pattern.Length())
	}
}
