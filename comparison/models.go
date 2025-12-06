package comparison

import "github.com/anthropics/anthropic-sdk-go"

// Model API identifiers (raw strings accepted by the SDK)
const (
	ModelHaikuAPI  = "claude-3-5-haiku-latest"
	ModelSonnetAPI = "claude-sonnet-4-20250514"
	ModelOpusAPI   = "claude-opus-4-5-20251101"
)

// ModelConfig represents configuration for an available AI model
type ModelConfig struct {
	ID          string          // Short identifier (e.g., "haiku", "sonnet", "opus")
	DisplayName string          // Human-readable name (e.g., "Haiku", "Sonnet", "Opus")
	APIModel    anthropic.Model // SDK model (string type)
	Provider    string          // AI provider (e.g., "anthropic")
}

// AvailableModels is the registry of all supported AI models
var AvailableModels = []ModelConfig{
	{
		ID:          "haiku",
		DisplayName: "Haiku",
		APIModel:    ModelHaikuAPI,
		Provider:    "anthropic",
	},
	{
		ID:          "sonnet",
		DisplayName: "Sonnet",
		APIModel:    ModelSonnetAPI,
		Provider:    "anthropic",
	},
	{
		ID:          "opus",
		DisplayName: "Opus",
		APIModel:    ModelOpusAPI,
		Provider:    "anthropic",
	},
}

// DefaultModel is the model used when no flag is provided
var DefaultModel = AvailableModels[0] // Haiku

// GetModelByID looks up a model configuration by its short identifier
func GetModelByID(id string) (*ModelConfig, bool) {
	for _, m := range AvailableModels {
		if m.ID == id {
			return &m, true
		}
	}
	return nil, false
}

// GetModelIDs returns a slice of all available model IDs
func GetModelIDs() []string {
	ids := make([]string, len(AvailableModels))
	for i, m := range AvailableModels {
		ids[i] = m.ID
	}
	return ids
}
