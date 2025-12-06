# Quickstart: Model Comparison Framework

**Date**: 2025-12-06
**Feature**: 003-model-comparison

## Prerequisites

- Interplay built and running
- `ANTHROPIC_API_KEY` environment variable set
- Basic familiarity with interplay commands

## Quick Examples

### 1. Compare Models on a Prompt

Run the same prompt against Haiku, Sonnet, and Opus:

```
> compare create a funky bass line
Running comparison: "create a funky bass line"
  [1/3] Haiku... done (1.2s)
  [2/3] Sonnet... done (2.3s)
  [3/3] Opus... done (3.1s)

Comparison saved: 2025-12-06T14-30-00_funky-bass
```

### 2. Load and Listen to Results

Load each model's pattern to hear the differences:

```
> compare-load 2025-12-06T14-30-00_funky-bass haiku
Loaded Haiku result...

# Listen to the pattern playing...

> compare-load 2025-12-06T14-30-00_funky-bass sonnet
Loaded Sonnet result...

# Compare the difference
```

### 3. Blind Evaluation (Unbiased)

Evaluate without knowing which model made which pattern:

```
> blind 2025-12-06T14-30-00_funky-bass
Entering blind evaluation mode...

Patterns available:
  A - 48 steps, 110 BPM
  B - 48 steps, 105 BPM
  C - 48 steps, 108 BPM

> load A
# Listen...

> rate A 4

> load B
# Listen...

> rate B 5

> load C
# Listen...

> rate C 3

> reveal
Results:
  A (rated 4) -> Haiku
  B (rated 5) -> Sonnet
  C (rated 3) -> Opus
```

### 4. Switch Default Model

Use a different model for regular AI interactions:

```bash
# Start with Sonnet as default
./interplay --model sonnet
```

Or switch during session:

```
> model sonnet
Switched to Sonnet

> ai create something dark
# Now uses Sonnet instead of Haiku
```

### 5. List Available Models

```
> models
Available AI models:
  haiku   - Haiku (fastest, cheapest) [active]
  sonnet  - Sonnet (balanced)
  opus    - Opus (most capable)
```

## Typical Workflow

1. **Explore**: Run `compare` with prompts you care about
2. **Listen**: Use `compare-load` to hear each result
3. **Evaluate**: Use `blind` mode for unbiased comparison
4. **Decide**: Choose your preferred default model with `model` command
5. **Clean up**: Delete old comparisons with `compare-delete`

## Tips

- **Cost awareness**: Opus is most expensive, Haiku is cheapest. Use comparisons strategically.
- **Blind mode**: Most useful when you suspect bias toward a particular model.
- **Ratings**: Even informal ratings help track which models work best for your style.
- **Pattern length**: Comparisons work best with similar-length prompts for fair comparison.

## Files Created

Comparisons are saved in:
```
comparisons/
├── 2025-12-06T14-30-00_funky-bass.json
├── 2025-12-05T10-15-00_dark-techno.json
└── ...
```

You can back up or share these files.
