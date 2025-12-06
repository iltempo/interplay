package comparison

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iltempo/interplay/sequence"
)

// ComparisonStatus represents the status of a comparison run
type ComparisonStatus string

const (
	StatusRunning   ComparisonStatus = "running"
	StatusComplete  ComparisonStatus = "complete"
	StatusPartial   ComparisonStatus = "partial"
	StatusCancelled ComparisonStatus = "cancelled"
)

// ResultStatus represents the status of a single model's result
type ResultStatus string

const (
	ResultSuccess    ResultStatus = "success"
	ResultError      ResultStatus = "error"
	ResultTimeout    ResultStatus = "timeout"
	ResultParseError ResultStatus = "parse_error"
)

// Comparison represents a single comparison run containing results from multiple AI models
type Comparison struct {
	ID        string                  `json:"id"`
	CreatedAt time.Time               `json:"created_at"`
	Prompt    string                  `json:"prompt"`
	Status    ComparisonStatus        `json:"status"`
	Results   []ModelResult           `json:"results"`
	Ratings   map[string]*Rating      `json:"ratings,omitempty"`
}

// ModelResult represents the output from a single model for a comparison
type ModelResult struct {
	Model            string                `json:"model"`              // API model identifier
	ModelDisplayName string                `json:"model_display_name"` // Human-readable name
	Status           ResultStatus          `json:"status"`
	Commands         []string              `json:"commands,omitempty"`
	Pattern          *sequence.PatternFile `json:"pattern,omitempty"`
	RawResponse      string                `json:"raw_response,omitempty"` // Stored on parse_error
	Error            string                `json:"error,omitempty"`
	DurationMs       int64                 `json:"duration_ms"`
}

// GetResultByModelID returns the result for a specific model, or nil if not found
func (c *Comparison) GetResultByModelID(modelID string) *ModelResult {
	for i := range c.Results {
		if c.Results[i].Model == modelID {
			return &c.Results[i]
		}
	}
	return nil
}

// GetResultByDisplayName returns the result for a model by display name, or nil if not found
func (c *Comparison) GetResultByDisplayName(displayName string) *ModelResult {
	for i := range c.Results {
		if c.Results[i].ModelDisplayName == displayName {
			return &c.Results[i]
		}
	}
	return nil
}

// SuccessfulResults returns only results with success status
func (c *Comparison) SuccessfulResults() []ModelResult {
	var results []ModelResult
	for _, r := range c.Results {
		if r.Status == ResultSuccess {
			results = append(results, r)
		}
	}
	return results
}

// HasRating checks if a model has been rated in this comparison
func (c *Comparison) HasRating(modelID string) bool {
	if c.Ratings == nil {
		return false
	}
	_, exists := c.Ratings[modelID]
	return exists
}

// ComparisonsDir is the directory where comparisons are saved
const ComparisonsDir = "comparisons"

// generateComparisonID creates a unique ID based on timestamp
func generateComparisonID() string {
	return time.Now().Format("20060102-150405")
}

// ProgressCallback is called during comparison execution to report progress
type ProgressCallback func(modelName string, status string)

// CommandGenerator is an interface for generating commands from a prompt
// This allows decoupling from the AI client
type CommandGenerator interface {
	GenerateCommandsWithModel(ctx context.Context, prompt string, pattern *sequence.Pattern, model string) ([]string, string, error)
}

// RunComparison executes the same prompt against all available models
// Returns a Comparison with results from each model
func RunComparison(ctx context.Context, prompt string, generator CommandGenerator, progress ProgressCallback) (*Comparison, error) {
	comparison := &Comparison{
		ID:        generateComparisonID(),
		CreatedAt: time.Now(),
		Prompt:    prompt,
		Status:    StatusRunning,
		Results:   make([]ModelResult, 0, len(AvailableModels)),
	}

	successCount := 0

	for _, modelConfig := range AvailableModels {
		if progress != nil {
			progress(modelConfig.DisplayName, "running")
		}

		result := executePromptForModel(ctx, prompt, generator, modelConfig)
		comparison.Results = append(comparison.Results, result)

		if result.Status == ResultSuccess {
			successCount++
		}

		if progress != nil {
			progress(modelConfig.DisplayName, string(result.Status))
		}
	}

	// Determine final status
	if successCount == len(AvailableModels) {
		comparison.Status = StatusComplete
	} else if successCount > 0 {
		comparison.Status = StatusPartial
	} else {
		comparison.Status = StatusCancelled
	}

	return comparison, nil
}

// executePromptForModel runs a single prompt against one model and captures the result
func executePromptForModel(ctx context.Context, prompt string, generator CommandGenerator, modelConfig ModelConfig) ModelResult {
	start := time.Now()

	result := ModelResult{
		Model:            string(modelConfig.APIModel),
		ModelDisplayName: modelConfig.DisplayName,
	}

	// Create a fresh pattern for this model's execution
	pattern := sequence.New(sequence.DefaultPatternLength)

	// Generate commands
	commands, rawResponse, err := generator.GenerateCommandsWithModel(ctx, prompt, pattern, string(modelConfig.APIModel))
	elapsed := time.Since(start)
	result.DurationMs = elapsed.Milliseconds()

	if err != nil {
		result.Status = ResultError
		result.Error = err.Error()
		return result
	}

	// Check if we got any commands
	if len(commands) == 0 {
		result.Status = ResultParseError
		result.RawResponse = rawResponse
		result.Error = "no commands generated"
		return result
	}

	result.Status = ResultSuccess
	result.Commands = commands

	// Execute commands to build the pattern
	for _, cmd := range commands {
		// Simple command execution - we just need to capture the pattern state
		// The actual execution happens in the command handler
		executePatternCommand(pattern, cmd)
	}

	// Convert pattern to PatternFile for storage
	patternFile := pattern.ToPatternFile(fmt.Sprintf("%s_%s", modelConfig.ID, prompt[:min(20, len(prompt))]))
	result.Pattern = patternFile

	return result
}

// executePatternCommand executes a single command on a pattern
// This is a simplified version that handles the core pattern commands
func executePatternCommand(p *sequence.Pattern, cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch strings.ToLower(parts[0]) {
	case "set":
		if len(parts) >= 3 {
			var step int
			fmt.Sscanf(parts[1], "%d", &step)
			note := parts[2]
			if strings.ToLower(note) == "rest" {
				p.SetRest(step)
			} else {
				midiNote, err := sequence.NoteNameToMIDI(note)
				if err == nil {
					p.SetNote(step, midiNote)
					// Parse optional parameters
					for i := 3; i < len(parts); i++ {
						param := parts[i]
						if strings.HasPrefix(param, "vel:") {
							var vel int
							fmt.Sscanf(strings.TrimPrefix(param, "vel:"), "%d", &vel)
							if vel >= 0 && vel <= 127 {
								p.SetVelocity(step, uint8(vel))
							}
						} else if strings.HasPrefix(param, "gate:") {
							var gate int
							fmt.Sscanf(strings.TrimPrefix(param, "gate:"), "%d", &gate)
							if gate >= 1 && gate <= 100 {
								p.SetGate(step, gate)
							}
						} else if strings.HasPrefix(param, "dur:") {
							var dur int
							fmt.Sscanf(strings.TrimPrefix(param, "dur:"), "%d", &dur)
							if dur >= 1 {
								p.SetNoteWithDuration(step, midiNote, dur)
							}
						}
					}
				}
			}
		}
	case "rest":
		if len(parts) >= 2 {
			var step int
			fmt.Sscanf(parts[1], "%d", &step)
			p.SetRest(step)
		}
	case "clear":
		p.Clear()
	case "tempo":
		if len(parts) >= 2 {
			var bpm int
			fmt.Sscanf(parts[1], "%d", &bpm)
			p.SetTempo(bpm)
		}
	case "length":
		if len(parts) >= 2 {
			var length int
			fmt.Sscanf(parts[1], "%d", &length)
			p.Resize(length)
		}
	case "velocity":
		if len(parts) >= 3 {
			var step, vel int
			fmt.Sscanf(parts[1], "%d", &step)
			fmt.Sscanf(parts[2], "%d", &vel)
			if vel >= 0 && vel <= 127 {
				p.SetVelocity(step, uint8(vel))
			}
		}
	case "gate":
		if len(parts) >= 3 {
			var step, gate int
			fmt.Sscanf(parts[1], "%d", &step)
			fmt.Sscanf(parts[2], "%d", &gate)
			p.SetGate(step, gate)
		}
	case "humanize":
		if len(parts) >= 3 {
			var amount int
			fmt.Sscanf(parts[2], "%d", &amount)
			switch strings.ToLower(parts[1]) {
			case "velocity", "vel":
				p.SetHumanizeVelocity(amount)
			case "timing", "time":
				p.SetHumanizeTiming(amount)
			case "gate":
				p.SetHumanizeGate(amount)
			}
		}
	case "swing":
		if len(parts) >= 2 {
			var swing int
			fmt.Sscanf(parts[1], "%d", &swing)
			p.SetSwing(swing)
		}
	}
}

// SaveComparison saves a comparison to a JSON file
func SaveComparison(c *Comparison) error {
	// Ensure comparisons directory exists
	if err := os.MkdirAll(ComparisonsDir, 0755); err != nil {
		return fmt.Errorf("failed to create comparisons directory: %w", err)
	}

	// Create file path using ID
	filename := c.ID + ".json"
	filePath := filepath.Join(ComparisonsDir, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal comparison: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write comparison file: %w", err)
	}

	return nil
}

// LoadComparison loads a comparison from a JSON file
func LoadComparison(id string) (*Comparison, error) {
	filename := id + ".json"
	filePath := filepath.Join(ComparisonsDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("comparison '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to read comparison file: %w", err)
	}

	var c Comparison
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse comparison file: %w", err)
	}

	return &c, nil
}

// ListComparisons returns a list of all saved comparison IDs
func ListComparisons() ([]string, error) {
	if _, err := os.Stat(ComparisonsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(ComparisonsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read comparisons directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			id := strings.TrimSuffix(entry.Name(), ".json")
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// DeleteComparison deletes a saved comparison
func DeleteComparison(id string) error {
	filename := id + ".json"
	filePath := filepath.Join(ComparisonsDir, filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("comparison '%s' not found", id)
		}
		return fmt.Errorf("failed to delete comparison: %w", err)
	}

	return nil
}
