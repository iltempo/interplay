package sequence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PatternsDir = "patterns"
)

// PatternStep represents a single step in the JSON format
type PatternStep struct {
	Step     int    `json:"step"`
	Note     string `json:"note"`
	Velocity uint8  `json:"velocity,omitempty"`
	Gate     int    `json:"gate,omitempty"`
}

// PatternFile represents the JSON structure for saving/loading patterns
type PatternFile struct {
	Name      string        `json:"name"`
	Tempo     int           `json:"tempo"`
	Steps     []PatternStep `json:"steps"`
	CreatedAt string        `json:"created_at,omitempty"`
}

// ToPatternFile converts a Pattern to the JSON-serializable format
func (p *Pattern) ToPatternFile(name string) *PatternFile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pf := &PatternFile{
		Name:      name,
		Tempo:     p.BPM,
		Steps:     make([]PatternStep, 0, NumSteps),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Only include non-rest steps
	for i := 0; i < NumSteps; i++ {
		if !p.Steps[i].IsRest {
			step := p.Steps[i]
			ps := PatternStep{
				Step: i + 1, // 1-indexed for user
				Note: midiToNoteName(step.Note),
			}
			// Only include velocity/gate if non-default
			if step.Velocity != 100 {
				ps.Velocity = step.Velocity
			}
			if step.Gate != 90 {
				ps.Gate = step.Gate
			}
			pf.Steps = append(pf.Steps, ps)
		}
	}

	return pf
}

// FromPatternFile creates a new Pattern from the JSON format
func FromPatternFile(pf *PatternFile) (*Pattern, error) {
	p := &Pattern{
		BPM: pf.Tempo,
	}

	// Initialize all steps as rests
	for i := 0; i < NumSteps; i++ {
		p.Steps[i] = Step{IsRest: true}
	}

	// Set notes from file
	for _, ps := range pf.Steps {
		if ps.Step < 1 || ps.Step > NumSteps {
			return nil, fmt.Errorf("invalid step number: %d", ps.Step)
		}

		midiNote, err := NoteNameToMIDI(ps.Note)
		if err != nil {
			return nil, fmt.Errorf("invalid note in step %d: %w", ps.Step, err)
		}

		// Use defaults if not specified in JSON
		velocity := ps.Velocity
		if velocity == 0 {
			velocity = 100
		}
		gate := ps.Gate
		if gate == 0 {
			gate = 90
		}

		p.Steps[ps.Step-1] = Step{
			Note:     midiNote,
			IsRest:   false,
			Velocity: velocity,
			Gate:     gate,
		}
	}

	return p, nil
}

// Save saves the pattern to a JSON file in the patterns directory
func (p *Pattern) Save(name string) error {
	// Ensure patterns directory exists
	if err := os.MkdirAll(PatternsDir, 0755); err != nil {
		return fmt.Errorf("failed to create patterns directory: %w", err)
	}

	// Convert to JSON format
	pf := p.ToPatternFile(name)

	// Create file path
	filename := sanitizeFilename(name) + ".json"
	filepath := filepath.Join(PatternsDir, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pattern: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write pattern file: %w", err)
	}

	return nil
}

// Load loads a pattern from a JSON file in the patterns directory
func Load(name string) (*Pattern, error) {
	// Create file path
	filename := sanitizeFilename(name) + ".json"
	filepath := filepath.Join(PatternsDir, filename)

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("pattern '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read pattern file: %w", err)
	}

	// Unmarshal JSON
	var pf PatternFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("failed to parse pattern file: %w", err)
	}

	// Convert to Pattern
	return FromPatternFile(&pf)
}

// List returns a list of all saved pattern names
func List() ([]string, error) {
	// Check if patterns directory exists
	if _, err := os.Stat(PatternsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(PatternsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns directory: %w", err)
	}

	// Collect .json files
	var patterns []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Remove .json extension
			name := strings.TrimSuffix(entry.Name(), ".json")
			patterns = append(patterns, name)
		}
	}

	return patterns, nil
}

// Delete deletes a saved pattern
func Delete(name string) error {
	// Create file path
	filename := sanitizeFilename(name) + ".json"
	filepath := filepath.Join(PatternsDir, filename)

	// Delete file
	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pattern '%s' not found", name)
		}
		return fmt.Errorf("failed to delete pattern: %w", err)
	}

	return nil
}

// sanitizeFilename removes potentially problematic characters from filenames
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	result := sb.String()
	if result == "" {
		return "unnamed"
	}
	return result
}
