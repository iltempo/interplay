# Data Model: Model Comparison Framework

**Date**: 2025-12-06
**Feature**: 003-model-comparison

## Entities

### Comparison

A single comparison run containing results from multiple AI models for the same prompt.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| ID | string | Unique identifier | Format: `YYYY-MM-DDTHH-MM-SS_<sanitized-prompt>` |
| CreatedAt | timestamp | When comparison was created | RFC3339 format |
| Prompt | string | The user's prompt sent to all models | Non-empty |
| Status | enum | Comparison status | `running`, `complete`, `partial`, `cancelled` |
| Results | []ModelResult | Results from each model | 1-N results |
| Ratings | map[string]Rating | User ratings keyed by model ID | Optional |

**State Transitions**:
```
[new] -> running -> complete
              \-> partial (some models failed)
              \-> cancelled (user cancelled)
```

### ModelResult

Output from a single model for a comparison.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| Model | string | Model API identifier | e.g., `claude-3-5-haiku-latest` |
| ModelDisplayName | string | Human-readable model name | e.g., `Haiku` |
| Status | enum | Result status | `success`, `error`, `timeout`, `parse_error` |
| Commands | []string | Generated commands | May be empty |
| Pattern | PatternFile | Resulting pattern state | Null if error |
| RawResponse | string | Raw AI response | Stored on parse_error |
| Error | string | Error message | Null if success |
| DurationMs | int64 | API call duration | Milliseconds |

### ModelConfig

Configuration for an available AI model.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| ID | string | Short identifier | e.g., `haiku`, `sonnet`, `opus` |
| DisplayName | string | Human-readable name | e.g., `Haiku`, `Sonnet`, `Opus` |
| APIModel | string | SDK model constant | Maps to `anthropic.Model` |
| Provider | string | AI provider | `anthropic` (extensible) |

**Available Models** (initial):
| ID | DisplayName | APIModel | Provider |
|----|-------------|----------|----------|
| haiku | Haiku | `claude-3-5-haiku-latest` | anthropic |
| sonnet | Sonnet | `claude-3-5-sonnet-latest` | anthropic |
| opus | Opus | `claude-3-5-opus-20241022` | anthropic |

### Rating

User's subjective assessment of a model's output.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| RhythmicInterest | int | Syncopation, variety, groove | 1-5 |
| VelocityDynamics | int | Accent placement, variation | 1-5 |
| GenreAccuracy | int | Matches requested style | 1-5 |
| Overall | int | Overall quality assessment | 1-5 |

### BlindSession

In-memory state for blind evaluation (not persisted).

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| ComparisonID | string | Reference to comparison | Must exist |
| LabelMapping | map[string]string | Label -> Model ID | e.g., `A -> haiku` |
| Rated | map[string]bool | Which labels have been rated | |
| Complete | bool | All patterns rated | |

## Relationships

```
Comparison 1---* ModelResult   (one comparison has many model results)
Comparison 1---* Rating        (one comparison can have ratings for multiple models)
ModelResult *---1 ModelConfig  (each result references one model config)
BlindSession *---1 Comparison  (blind session references one comparison)
```

## Storage Schema

### comparisons/<id>.json

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
      "commands": [
        "clear",
        "tempo 110",
        "set 1 E2 vel:120 dur:2",
        "set 4 G2 vel:85"
      ],
      "pattern": {
        "name": "comparison_haiku_2025-12-06T14-30-00",
        "tempo": 110,
        "length": 48,
        "steps": [
          {"step": 1, "note": "E2", "velocity": 120, "duration": 2},
          {"step": 4, "note": "G2", "velocity": 85}
        ]
      },
      "raw_response": null,
      "error": null,
      "duration_ms": 1234
    },
    {
      "model": "claude-3-5-sonnet-latest",
      "model_display_name": "Sonnet",
      "status": "success",
      "commands": [
        "clear",
        "tempo 105",
        "set 1 C3 vel:100 dur:4"
      ],
      "pattern": {
        "name": "comparison_sonnet_2025-12-06T14-30-00",
        "tempo": 105,
        "length": 48,
        "steps": [
          {"step": 1, "note": "C3", "velocity": 100, "duration": 4}
        ]
      },
      "raw_response": null,
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
    },
    "claude-3-5-sonnet-latest": {
      "rhythmic_interest": 5,
      "velocity_dynamics": 4,
      "genre_accuracy": 5,
      "overall": 5
    }
  }
}
```

## Validation Rules

### Comparison
- ID must be unique within `comparisons/` directory
- Prompt must be non-empty string
- Status must be valid enum value
- At least one result required

### ModelResult
- Model must be a known model identifier
- If status is `success`, Pattern must be non-null
- If status is `error` or `timeout`, Error must be non-null
- If status is `parse_error`, RawResponse should be preserved

### Rating
- All scores must be integers 1-5
- Can only rate models present in comparison results

### BlindSession
- ComparisonID must reference existing comparison
- LabelMapping must cover all models in comparison
- Labels are single uppercase letters (A, B, C, ...)

## Index / Lookup Patterns

1. **List comparisons**: Read all `*.json` files in `comparisons/` directory, extract ID and metadata
2. **Get comparison by ID**: Direct file access `comparisons/<id>.json`
3. **Get available models**: In-memory registry lookup
4. **Get model by ID**: Registry lookup with ID key
