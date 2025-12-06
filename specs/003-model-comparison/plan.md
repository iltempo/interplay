# Implementation Plan: Model Comparison Framework

**Branch**: `003-model-comparison` | **Date**: 2025-12-06 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-model-comparison/spec.md`

## Summary

Build a model comparison framework that allows users to systematically compare AI model outputs (Haiku, Sonnet, Opus) for pattern generation quality. The framework will execute the same prompt against multiple models, save results for later review, support blind evaluation mode for unbiased assessment, and allow rating patterns on musical criteria.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**: `anthropic-sdk-go` (existing), `gitlab.com/gomidi/midi/v2` (existing)
**Storage**: JSON files in `comparisons/` directory (following existing `patterns/` convention)
**Testing**: Standard Go testing (`go test ./...`)
**Target Platform**: macOS/Windows/Linux CLI application
**Project Type**: Single CLI application (existing structure)
**Performance Goals**: Comparison results saved within 60 seconds per model, load/list within 2 seconds
**Constraints**: Must integrate with existing `ai/` module, maintain backward compatibility
**Scale/Scope**: Local single-user tool, ~10-100 saved comparisons typical

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Incremental Development | PASS | Feature adds new module without modifying core playback/sequence |
| II. Collaborative Decision-Making | PASS | Spec reviewed and clarified with user before planning |
| III. Musical Intelligence | PASS | Supports evaluating AI musical output quality |
| IV. Pattern-Based Simplicity | PASS | Comparisons saved independently, patterns loaded via existing mechanism |
| V. Learning-First Documentation | PASS | Will document new commands and workflow |
| VI. AI-First Creativity | PASS | Directly supports improving AI model selection for creative work |

**Architecture Constraints Check**:
- Core modules remain unchanged (PASS)
- New `comparison/` module follows existing module pattern (PASS)
- JSON storage in `comparisons/` directory (PASS)
- AI integration via existing `ai/` module with model switching (PASS)

**No violations requiring justification.**

## Project Structure

### Documentation (this feature)

```text
specs/003-model-comparison/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI command specs)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
# Existing structure (unchanged)
ai/
├── ai.go                # Existing AI client - add model switching
└── ai_test.go

commands/
├── commands.go          # Add new comparison commands
└── ...

sequence/
├── sequence.go          # Existing pattern management
├── persistence.go       # Existing JSON save/load
└── ...

# New module for this feature
comparison/
├── comparison.go        # Comparison entity, storage, listing
├── comparison_test.go
├── rating.go            # Rating entity and persistence
├── blind.go             # Blind evaluation session management
└── models.go            # Model configuration registry

# New directory for comparison data
comparisons/             # JSON files for saved comparisons
```

**Structure Decision**: Add new `comparison/` module following existing module pattern. Extends `ai/` module for model switching. Adds `comparisons/` data directory sibling to `patterns/`.

## Complexity Tracking

> No violations - table not required.
