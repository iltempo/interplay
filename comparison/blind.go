package comparison

import (
	"crypto/rand"
	"math/big"
)

// BlindSession represents an active blind evaluation session (in-memory, not persisted)
type BlindSession struct {
	ComparisonID string            // Reference to comparison being evaluated
	LabelMapping map[string]string // Label -> Model API ID (e.g., "A" -> "claude-3-5-haiku-latest")
	ModelToLabel map[string]string // Model API ID -> Label (reverse lookup)
	Rated        map[string]int    // Label -> Rating score (overall only for blind mode)
	Labels       []string          // Ordered list of labels (A, B, C, ...)
}

// NewBlindSession creates a new blind evaluation session with randomized label assignment
func NewBlindSession(comparisonID string, modelIDs []string) *BlindSession {
	session := &BlindSession{
		ComparisonID: comparisonID,
		LabelMapping: make(map[string]string),
		ModelToLabel: make(map[string]string),
		Rated:        make(map[string]int),
		Labels:       make([]string, len(modelIDs)),
	}

	// Create shuffled copy of model IDs
	shuffled := make([]string, len(modelIDs))
	copy(shuffled, modelIDs)
	shuffleStrings(shuffled)

	// Assign labels A, B, C, ... to shuffled models
	for i, modelID := range shuffled {
		label := string(rune('A' + i))
		session.Labels[i] = label
		session.LabelMapping[label] = modelID
		session.ModelToLabel[modelID] = label
	}

	return session
}

// GetModelIDByLabel returns the model API ID for a given label
func (s *BlindSession) GetModelIDByLabel(label string) (string, bool) {
	modelID, exists := s.LabelMapping[label]
	return modelID, exists
}

// GetLabelByModelID returns the label for a given model API ID
func (s *BlindSession) GetLabelByModelID(modelID string) (string, bool) {
	label, exists := s.ModelToLabel[modelID]
	return label, exists
}

// RateLabel records a rating for a label
func (s *BlindSession) RateLabel(label string, score int) bool {
	if _, exists := s.LabelMapping[label]; !exists {
		return false
	}
	s.Rated[label] = score
	return true
}

// GetRating returns the rating for a label, or 0 if not rated
func (s *BlindSession) GetRating(label string) int {
	return s.Rated[label]
}

// IsRated checks if a label has been rated
func (s *BlindSession) IsRated(label string) bool {
	_, exists := s.Rated[label]
	return exists
}

// IsComplete checks if all labels have been rated
func (s *BlindSession) IsComplete() bool {
	return len(s.Rated) == len(s.Labels)
}

// RatedCount returns the number of patterns that have been rated
func (s *BlindSession) RatedCount() int {
	return len(s.Rated)
}

// TotalCount returns the total number of patterns to rate
func (s *BlindSession) TotalCount() int {
	return len(s.Labels)
}

// Reveal returns the label-to-model mapping with ratings
type RevealResult struct {
	Label       string
	ModelID     string
	DisplayName string
	Rating      int
}

// GetRevealResults returns the reveal results (call after IsComplete)
func (s *BlindSession) GetRevealResults(comparison *Comparison) []RevealResult {
	results := make([]RevealResult, len(s.Labels))
	for i, label := range s.Labels {
		modelID := s.LabelMapping[label]
		displayName := ""
		for _, r := range comparison.Results {
			if r.Model == modelID {
				displayName = r.ModelDisplayName
				break
			}
		}
		results[i] = RevealResult{
			Label:       label,
			ModelID:     modelID,
			DisplayName: displayName,
			Rating:      s.Rated[label],
		}
	}
	return results
}

// shuffleStrings performs a Fisher-Yates shuffle using crypto/rand
func shuffleStrings(slice []string) {
	for i := len(slice) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// Fallback: no shuffle on error (shouldn't happen)
			return
		}
		j := int(jBig.Int64())
		slice[i], slice[j] = slice[j], slice[i]
	}
}
