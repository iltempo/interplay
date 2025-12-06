# CLI Commands Contract: Model Comparison Framework

**Date**: 2025-12-06
**Feature**: 003-model-comparison

This document specifies the CLI commands added by the Model Comparison Framework.

## Command-Line Flags

### --model

Select AI model at startup.

```
./interplay --model <model-id>
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| model-id | string | No | `haiku` | Model identifier: `haiku`, `sonnet`, `opus` |

**Examples**:
```bash
./interplay --model sonnet    # Start with Sonnet as default AI model
./interplay --model opus      # Start with Opus as default AI model
./interplay                   # Uses Haiku (default)
```

**Errors**:
- Unknown model ID: `Error: unknown model 'xyz'. Available: haiku, sonnet, opus`

---

## Runtime Commands

### model

Switch the active AI model during session.

```
model <model-id>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| model-id | string | Yes | Model identifier: `haiku`, `sonnet`, `opus` |

**Output**:
```
Switched to Sonnet (claude-3-5-sonnet-latest)
```

**Errors**:
- Missing argument: `Usage: model <model-id>`
- Unknown model: `Error: unknown model 'xyz'. Available: haiku, sonnet, opus`

---

### models

List available AI models.

```
models
```

**Output**:
```
Available AI models:
  haiku   - Haiku (claude-3-5-haiku-latest) [active]
  sonnet  - Sonnet (claude-3-5-sonnet-latest)
  opus    - Opus (claude-3-5-opus-20241022)
```

---

### compare

Run comparison against all models with a prompt.

```
compare <prompt>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| prompt | string | Yes | Musical prompt to send to all models |

**Output** (during execution):
```
Running comparison: "create a funky bass line"
  [1/3] Haiku... done (1.2s)
  [2/3] Sonnet... done (2.3s)
  [3/3] Opus... done (3.1s)

Comparison saved: 2025-12-06T14-30-00_funky-bass

Results summary:
  Haiku:  8 commands, tempo 110, 12 notes
  Sonnet: 12 commands, tempo 105, 16 notes
  Opus:   15 commands, tempo 108, 18 notes

Use 'compare-view 2025-12-06T14-30-00_funky-bass' to see details
Use 'compare-load 2025-12-06T14-30-00_funky-bass haiku' to load a result
```

**Output** (on model failure):
```
  [2/3] Sonnet... error: API rate limit exceeded
```

**Errors**:
- Missing prompt: `Usage: compare <prompt>`
- No API key: `Error: ANTHROPIC_API_KEY not set`
- All models failed: `Error: all models failed. Check your API key and network.`

---

### compare-list

List all saved comparisons.

```
compare-list
```

**Output**:
```
Saved comparisons:
  2025-12-06T14-30-00_funky-bass     "create a funky bass line"      3 models
  2025-12-05T10-15-00_dark-techno    "make a dark techno pattern"    3 models
  2025-12-04T16-45-00_jazz-groove    "create a jazz groove"          2 models (partial)

Total: 3 comparisons
```

**Output** (no comparisons):
```
No saved comparisons found.
```

---

### compare-view

View details of a saved comparison.

```
compare-view <id>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comparison ID |

**Output**:
```
Comparison: 2025-12-06T14-30-00_funky-bass
Created: 2025-12-06 14:30:00
Prompt: "create a funky bass line"
Status: complete

Results:

[Haiku] (1.2s)
Commands:
  clear
  tempo 110
  set 1 E2 vel:120 dur:2
  set 4 G2 vel:85
Pattern: 48 steps, 110 BPM, 4 notes
Rating: rhythmic=4 dynamics=3 genre=4 overall=4

[Sonnet] (2.3s)
Commands:
  clear
  tempo 105
  set 1 C3 vel:100 dur:4
  ...
Pattern: 48 steps, 105 BPM, 8 notes
Rating: (not rated)

[Opus] (3.1s)
Commands:
  ...
Pattern: 48 steps, 108 BPM, 10 notes
Rating: (not rated)
```

**Errors**:
- Missing ID: `Usage: compare-view <id>`
- Not found: `Error: comparison '2025-12-06T14-30-00_xyz' not found`

---

### compare-load

Load a pattern from a comparison result.

```
compare-load <id> <model-id>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comparison ID |
| model-id | string | Yes | Model identifier |

**Output**:
```
Loaded Haiku result from comparison '2025-12-06T14-30-00_funky-bass'
Pattern: 48 steps, 110 BPM, 4 notes
```

**Errors**:
- Missing arguments: `Usage: compare-load <id> <model-id>`
- Comparison not found: `Error: comparison '...' not found`
- Model not in comparison: `Error: model 'opus' not found in comparison`
- Model failed: `Error: model 'sonnet' failed in this comparison, no pattern to load`

---

### compare-delete

Delete a saved comparison.

```
compare-delete <id>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comparison ID |

**Output**:
```
Deleted comparison '2025-12-06T14-30-00_funky-bass'
```

**Errors**:
- Missing ID: `Usage: compare-delete <id>`
- Not found: `Error: comparison '...' not found`

---

### compare-rate

Rate a model's output in a comparison.

```
compare-rate <id> <model-id> <criteria> <score>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comparison ID |
| model-id | string | Yes | Model identifier |
| criteria | string | Yes | Rating criteria: `rhythmic`, `dynamics`, `genre`, `overall` |
| score | int | Yes | Score 1-5 |

**Output**:
```
Rated Haiku rhythmic_interest: 4
```

**Shorthand**: Rate all criteria at once:
```
compare-rate <id> <model-id> all <score>
```

**Errors**:
- Missing arguments: `Usage: compare-rate <id> <model-id> <criteria> <score>`
- Invalid criteria: `Error: invalid criteria 'xyz'. Use: rhythmic, dynamics, genre, overall, all`
- Invalid score: `Error: score must be 1-5`
- Model not found: `Error: model 'opus' not found in comparison`

---

### blind

Enter blind evaluation mode for a comparison.

```
blind <id>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Comparison ID |

**Output** (entering blind mode):
```
Entering blind evaluation mode for '2025-12-06T14-30-00_funky-bass'
Prompt: "create a funky bass line"

Patterns available (randomized order):
  A - 48 steps, 110 BPM
  B - 48 steps, 105 BPM
  C - 48 steps, 108 BPM

Commands in blind mode:
  load <label>           - Load pattern (e.g., 'load A')
  rate <label> <score>   - Rate overall 1-5 (e.g., 'rate A 4')
  reveal                 - End blind mode and show results
  exit                   - Exit without saving ratings

>
```

**Blind mode sub-commands**:

| Command | Description |
|---------|-------------|
| `load <label>` | Load pattern by label (A, B, C) |
| `rate <label> <score>` | Rate pattern overall (1-5) |
| `reveal` | End blind mode, show label-to-model mapping, save ratings |
| `exit` | Exit blind mode without saving |

**Output** (reveal):
```
Blind evaluation complete!

Results:
  A (rated 4) -> Haiku
  B (rated 5) -> Sonnet
  C (rated 3) -> Opus

Ratings saved to comparison.
```

**Errors**:
- Missing ID: `Usage: blind <id>`
- Not found: `Error: comparison '...' not found`
- Single model: `Warning: only 1 model in comparison, blind mode not useful. Continue? (y/n)`

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid arguments, file not found) |
| 2 | API error (network, authentication) |
