# Research: Model Comparison Framework

**Date**: 2025-12-06
**Feature**: 003-model-comparison

## Research Questions Resolved

### 1. Model Switching in Anthropic SDK

**Question**: How to dynamically switch between Claude models (Haiku, Sonnet, Opus) in the existing AI client?

**Decision**: Add a `model` field to the `ai.Client` struct with getter/setter methods. The model constant is passed to `MessageNewParams.Model` on each API call.

**Rationale**: The Anthropic SDK already supports model selection per-request via `anthropic.MessageNewParams{Model: ...}`. Current code hardcodes `anthropic.ModelClaude3_5HaikuLatest`. Simple change to make this configurable.

**Alternatives considered**:
- Create separate client instances per model: Rejected - wasteful, more complex lifecycle management
- Environment variable per model: Rejected - doesn't support runtime switching

**Implementation**:
```go
// Available model constants from anthropic SDK:
// - anthropic.ModelClaude3_5HaikuLatest  (default)
// - anthropic.ModelClaude35SonnetLatest
// - anthropic.ModelClaude3_5Opus20241022

type Client struct {
    client              anthropic.Client
    model               anthropic.Model  // NEW: configurable model
    conversationHistory []anthropic.MessageParam
}

func (c *Client) SetModel(model anthropic.Model) { c.model = model }
func (c *Client) GetModel() anthropic.Model { return c.model }
```

### 2. Comparison Storage Format

**Question**: What JSON structure should be used for comparison results?

**Decision**: Create a new JSON schema for comparisons, distinct from patterns. Store in `comparisons/` directory with timestamped filenames.

**Rationale**: Comparisons have different structure than patterns (multiple results per comparison, metadata about models, prompt text). Keeping them separate maintains clarity and allows independent schema evolution.

**Alternatives considered**:
- Embed in pattern files: Rejected - conflates concerns, bloats pattern files
- Database (SQLite): Rejected - overkill for local CLI tool, adds dependency

**Schema**:
```json
{
  "id": "2025-12-06T14-30-00_funky-bass",
  "created_at": "2025-12-06T14:30:00Z",
  "prompt": "create a funky bass line",
  "status": "complete",
  "results": [
    {
      "model": "claude-3-5-haiku-latest",
      "model_display_name": "Haiku",
      "status": "success",
      "commands": ["clear", "tempo 110", "set 1 E2 vel:120"],
      "pattern": { /* PatternFile structure */ },
      "error": null,
      "duration_ms": 1234
    },
    {
      "model": "claude-3-5-sonnet-latest",
      "model_display_name": "Sonnet",
      "status": "success",
      "commands": ["clear", "tempo 105", "set 1 C3 vel:100"],
      "pattern": { /* PatternFile structure */ },
      "error": null,
      "duration_ms": 2345
    }
  ],
  "ratings": {
    "claude-3-5-haiku-latest": {
      "rhythmic_interest": 4,
      "velocity_dynamics": 3,
      "genre_accuracy": 4,
      "overall": 4
    }
  }
}
```

### 3. Blind Evaluation Session State

**Question**: How to manage blind evaluation session state (label mapping, completion tracking)?

**Decision**: In-memory session state with transient mapping. Not persisted between runs.

**Rationale**: Blind sessions are short-lived evaluation workflows. User starts blind mode, evaluates, rates, sees reveal, done. No need to persist incomplete sessions.

**Alternatives considered**:
- Persist session to disk: Rejected - adds complexity, unclear use case for resuming blind sessions
- Store mapping in comparison file: Rejected - would reveal model identity if user inspects file

**Implementation**:
```go
type BlindSession struct {
    ComparisonID string
    LabelMapping map[string]string  // "A" -> "claude-3-5-haiku-latest"
    Rated        map[string]bool    // "A" -> true (has been rated)
    Complete     bool
}
```

### 4. Model Registry Pattern

**Question**: How to maintain list of available models for comparison and selection?

**Decision**: Hardcoded registry of supported Anthropic models in `comparison/models.go`. Registry returns model configs with ID, display name, and SDK constant.

**Rationale**: Initial scope is Anthropic-only. Hardcoded registry is simple and sufficient. Future extensibility can add provider abstraction when needed.

**Alternatives considered**:
- Configuration file: Rejected - unnecessary indirection for known models
- Environment-based: Rejected - models are fixed by SDK support

**Implementation**:
```go
type ModelConfig struct {
    ID          string          // "haiku", "sonnet", "opus"
    DisplayName string          // "Haiku", "Sonnet", "Opus"
    APIModel    anthropic.Model // anthropic.ModelClaude3_5HaikuLatest
    Provider    string          // "anthropic"
}

var AvailableModels = []ModelConfig{
    {ID: "haiku", DisplayName: "Haiku", APIModel: anthropic.ModelClaude3_5HaikuLatest, Provider: "anthropic"},
    {ID: "sonnet", DisplayName: "Sonnet", APIModel: anthropic.ModelClaude35SonnetLatest, Provider: "anthropic"},
    {ID: "opus", DisplayName: "Opus", APIModel: anthropic.ModelClaude3_5Opus20241022, Provider: "anthropic"},
}
```

### 5. Command Integration Pattern

**Question**: How to integrate new comparison commands with existing command structure?

**Decision**: Add new commands to `commands/commands.go` following existing pattern. Comparison-specific logic lives in `comparison/` module.

**Rationale**: Maintains consistency with existing codebase. Commands package handles parsing and dispatch; domain modules handle logic.

**Commands to add**:
- `compare <prompt>` - Run comparison against all models
- `compare-list` - List saved comparisons
- `compare-view <id>` - View comparison details
- `compare-load <id> <model>` - Load pattern from comparison result
- `compare-delete <id>` - Delete saved comparison
- `compare-rate <id> <model> <criteria> <score>` - Rate a model's output
- `blind <id>` - Enter blind evaluation mode
- `model <name>` - Switch active AI model
- `models` - List available models

### 6. Error Handling for Model Failures

**Question**: How to handle partial failures during multi-model comparison?

**Decision**: Continue with available models, mark failed models with error status and message in results.

**Rationale**: User wants to see what worked, not abort entire comparison on one failure. Matches spec FR-013.

**Implementation**:
```go
type ModelResult struct {
    // ...
    Status string  // "success", "error", "timeout"
    Error  string  // Error message if Status != "success"
}
```

## Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `anthropic-sdk-go` | existing | AI API calls |
| `encoding/json` | stdlib | Comparison persistence |
| `crypto/rand` | stdlib | Blind session label randomization |
| `time` | stdlib | Timestamps, duration tracking |

## Best Practices Applied

1. **Separation of concerns**: `comparison/` module handles comparison logic; `ai/` module handles API calls
2. **Existing patterns**: Follow `sequence/persistence.go` JSON patterns for comparison storage
3. **Graceful degradation**: Partial results saved on model failures
4. **Testability**: Pure functions for JSON serialization, mockable AI client interface
