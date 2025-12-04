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
	Step     int            `json:"step"`
	Note     string         `json:"note"`
	Velocity uint8          `json:"velocity,omitempty"`
	Gate     int            `json:"gate,omitempty"`
	Duration int            `json:"duration,omitempty"`
	CC       map[string]int `json:"cc,omitempty"` // CC automation: "74" -> 127 (JSON keys are strings)
}

// PatternFile represents the JSON structure for saving/loading patterns
type PatternFile struct {
	Name      string        `json:"name"`
	Tempo     int           `json:"tempo"`
	Length    int           `json:"length"`
	Steps     []PatternStep `json:"steps"`
	CreatedAt string        `json:"created_at,omitempty"`
}

// ToPatternFile converts a Pattern to the JSON-serializable format
func (p *Pattern) ToPatternFile(name string) *PatternFile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	patternLen := len(p.Steps)
	pf := &PatternFile{
		Name:      name,
		Tempo:     p.BPM,
		Length:    patternLen,
		Steps:     make([]PatternStep, 0, patternLen),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Only include non-rest steps
	for i := 0; i < patternLen; i++ {
		if !p.Steps[i].IsRest {
			step := p.Steps[i]
			ps := PatternStep{
				Step: i + 1, // 1-indexed for user
				Note: midiToNoteName(step.Note),
			}
			// Only include velocity/gate/duration if non-default
			if step.Velocity != 100 {
				ps.Velocity = step.Velocity
			}
			if step.Gate != 90 {
				ps.Gate = step.Gate
			}
			if step.Duration != 1 {
				ps.Duration = step.Duration
			}
			// Include CC automation if present (convert int keys to string keys for JSON)
			if len(step.CCValues) > 0 {
				ps.CC = make(map[string]int)
				for ccNum, value := range step.CCValues {
					ps.CC[fmt.Sprintf("%d", ccNum)] = value
				}
			}
			pf.Steps = append(pf.Steps, ps)
		}
	}

	return pf
}

// FromPatternFile creates a new Pattern from the JSON format
func FromPatternFile(pf *PatternFile) (*Pattern, error) {
	// Use the length from the file, or default if it's missing/invalid
	length := pf.Length
	if length <= 0 {
		// For backward compatibility with old files that have no length field,
		// we can try to infer it, or just default to 16.
		// For now, let's just use the default.
		length = DefaultPatternLength
	}

	p := New(length)
	p.BPM = pf.Tempo

	// Set notes from file
	for _, ps := range pf.Steps {
		if ps.Step < 1 || ps.Step > length {
			// Don't error out, just log it. This makes it robust to length changes.
			fmt.Printf("warning: step %d in pattern '%s' is out of bounds (length is %d), skipping\n", ps.Step, pf.Name, length)
			continue
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
		duration := ps.Duration
		if duration == 0 {
			duration = 1
		}

		// Convert CC map from JSON (string keys) to internal format (int keys)
		var ccValues map[int]int
		if len(ps.CC) > 0 {
			ccValues = make(map[int]int)
			for ccNumStr, value := range ps.CC {
				var ccNum int
				_, err := fmt.Sscanf(ccNumStr, "%d", &ccNum)
				if err != nil {
					fmt.Printf("warning: invalid CC number '%s' in step %d of pattern '%s', skipping\n", ccNumStr, ps.Step, pf.Name)
					continue
				}
				// Validate CC number and value
				if err := ValidateCC(ccNum, value); err != nil {
					fmt.Printf("warning: %v in step %d of pattern '%s', skipping\n", err, ps.Step, pf.Name)
					continue
				}
				ccValues[ccNum] = value
			}
		}

		p.Steps[ps.Step-1] = Step{
			Note:     midiNote,
			IsRest:   false,
			Velocity: velocity,
			Gate:     gate,
			Duration: duration,
			CCValues: ccValues,
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
