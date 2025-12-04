# Implementation Plan: MIDI CC Parameter Control

**Branch**: `001-midi-cc-control` | **Date**: 2025-12-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-midi-cc-control/spec.md`

## Summary

Add generic MIDI CC (Control Change) parameter control to Interplay, enabling users to control synthesizer parameters (filter, resonance, envelope, etc.) through both global commands and per-step automation. This foundational feature supports creative sound design and prepares for future AI-powered parameter suggestions (Phase 4b/4c).

**Key Capabilities:**
- Send any CC message (0-127) with any value (0-127) to connected MIDI synthesizers
- Per-step CC automation for dynamic parameter modulation (e.g., filter sweeps)
- Pattern persistence with full CC data fidelity (JSON save/load)
- Conversion command (`cc-apply`) to promote experimentation to permanent automation
- AI integration with updated system prompts for AI-generated CC automation

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**:
- `gitlab.com/gomidi/midi/v2` v2.3.16 (MIDI library)
- `github.com/anthropics/anthropic-sdk-go` v1.19.0 (AI integration)
- `github.com/chzyer/readline` v1.5.1 (command-line interface)

**Storage**: JSON files in `patterns/` directory (existing pattern persistence system)
**Testing**: `go test ./...` (standard Go testing, manual MIDI hardware testing)
**Target Platform**: macOS, Windows, Linux (platform-specific binaries due to CGO in rtmididrv)
**Project Type**: Single CLI application (monorepo structure)
**Performance Goals**:
- CC message timing within ±5ms of note messages at step boundaries
- No playback disruption when setting CC values
- Real-time CC parameter updates within one loop iteration (< 2 seconds at 80 BPM)

**Constraints**:
- Pattern-based loop boundary synchronization (no immediate mid-pattern updates)
- Standard MIDI CC specification (controller 0-127, value 0-127)
- Backward compatibility with existing pattern JSON files
- Thread-safe CC state management (mutex-protected shared state)

**Scale/Scope**:
- Single user, local MIDI device control
- 16-64 steps per pattern (variable length)
- Unlimited CC parameters per step
- Human-readable JSON persistence

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Incremental Development ✅ PASS

- Feature broken into 4 priority levels (P1-P4) for incremental delivery
- P1 (global CC) delivers immediate value and validates MIDI implementation
- Each priority independently testable and deployable
- Go concepts will be explained as introduced (maps for CC storage, etc.)

### Principle II: Collaborative Decision-Making ✅ PASS

- Clarification session resolved global vs per-step persistence semantics
- User approved transient global CC with `cc-apply` conversion approach
- Implementation awaits architectural discussion before coding

### Principle III: Musical Intelligence with Creative Freedom ✅ PASS

- CC automation maintains ±5ms timing precision (same as notes)
- Supports creative parameter control (filter, resonance, dissonance-inducing settings)
- Per-step automation enables musical expression through sound design
- No blocking of playback goroutine (CC messages sent at step boundaries)

### Principle IV: Pattern-Based Simplicity ✅ PASS

- CC changes queue and apply at loop boundaries (consistent with notes, velocity, gate)
- No immediate real-time CC updates mid-pattern
- Thread-safe CC state with mutex protection
- Background playback goroutine remains simple (just adds CC message sending)

### Principle V: Learning-First Documentation ✅ PASS

- Feature fully specified with user scenarios, requirements, edge cases
- Design decisions documented (global vs per-step, transient vs persistent)
- Quickstart guide will be created in Phase 1
- CLAUDE.md already updated with Phase 4a architecture

### Principle VI: AI-First Creativity ✅ PASS

- AI system prompts will be updated to include all CC commands
- AI can generate patterns with CC automation after implementation
- Foundation for Phase 4c AI sound design intelligence
- Conversion command (`cc-apply`) supports AI-assisted workflow

### Architectural Constraints ✅ PASS

- Extends existing core modules: `midi/` (CC messages), `sequence/` (CC data model), `playback/` (CC sending)
- Uses existing command parser framework: `commands/` module
- Pattern state mutex-protected (add CC values to existing pattern struct)
- JSON persistence extends existing save/load system

### Musical Constraints ✅ PASS

- Works within existing 16th-note timing resolution
- Compatible with humanization and swing (CC messages at step boundaries)
- Pattern length 1-64 steps supported (variable length compatible)
- JSON format remains human-readable and manually editable

**Gate Status**: ✅ ALL GATES PASS - Proceed to Phase 0 research

## Project Structure

### Documentation (this feature)

```text
specs/001-midi-cc-control/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (complete)
├── research.md          # Phase 0 output (to be generated)
├── data-model.md        # Phase 1 output (to be generated)
├── quickstart.md        # Phase 1 output (to be generated)
├── contracts/           # Phase 1 output (N/A for CLI - no external API contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
# Existing structure (to be extended)
/
├── ai/                  # AI integration - UPDATE system prompts
├── commands/            # Command parser - ADD cc, cc-step, cc-apply, cc-clear, cc-show
├── midi/                # MIDI communication - ADD CC message sending
├── playback/            # Pattern playback - ADD CC message dispatch at steps
├── sequence/            # Pattern state - ADD CC automation storage
├── patterns/            # JSON pattern files - EXTEND format with CC data
└── main.go              # Application entry point - NO CHANGES NEEDED
```

**Structure Decision**: Single project structure maintained. All changes extend existing modules without creating new top-level packages. This aligns with constitutional simplicity and incremental development principles.

## Complexity Tracking

> **No violations** - all constitutional gates pass without requiring complexity justifications.

## Planning Completion Summary

**Status**: ✅ PLANNING COMPLETE

### Deliverables Created

✅ **Phase 0: Research**
- `research.md` - 5 key design decisions documented with rationales
- Risk assessment completed (all risks mitigated)
- Go implementation patterns identified
- Testing strategy defined

✅ **Phase 1: Design & Contracts**
- `data-model.md` - Complete data structures and relationships
- `quickstart.md` - User-facing documentation with examples
- Agent context updated in `CLAUDE.md`
- No external API contracts needed (CLI application)

### Key Design Outcomes

**Data Model:**
- Extended `Step` struct with `CCValues map[int]int`
- Extended `Sequence` with `globalCC map[int]int`
- JSON format with backward-compatible `omitempty` tags

**User Workflow:**
- Global CC for experimentation (transient)
- Per-step CC for automation (persistent)
- `cc-apply` command for conversion
- Save warning prevents data loss

**Technical Decisions:**
- CC messages sent at step boundaries before Note On
- Mutex-protected state for thread safety
- Timing precision: ±5ms (same as notes)
- 100% backward compatible with existing patterns

### Next Steps

Ready to proceed to **Phase 2: Task Generation**

Run `/speckit.tasks` to generate implementation task breakdown (`tasks.md`).

**Command:**
```
/speckit.tasks
```

This will create dependency-ordered implementation tasks based on the completed design artifacts.

