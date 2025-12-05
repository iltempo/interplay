# Implementation Plan: Batch/Script Mode for Command Execution

**Branch**: `002-batch-script-mode` | **Date**: 2025-12-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-batch-script-mode/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add batch/script mode execution for Interplay commands via stdin piping and `--script` flag. This enables users to create reusable script files for performance setup (load patterns, configure settings) and testing automation. Scripts execute sequentially with pre-validation, graceful error handling, and real-time progress feedback. The application continues running with playback loop active after script completion unless an explicit `exit` command is present. This is a performance tool enhancement, not just batch processing.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**:
- `github.com/mattn/go-isatty` (cross-platform terminal detection for stdin mode, supports Cygwin/Git Bash)
- `flag` package (stdlib, command-line parsing)
- `bufio` package (stdlib, line-by-line reading)
- Existing: `gitlab.com/gomidi/midi/v2`, `anthropic-sdk-go`

**Storage**: File system (`patterns/` directory for JSON pattern files)
**Testing**: Standard Go testing (`go test ./...`), manual testing with script files
**Target Platform**: Cross-platform CLI (macOS, Linux, Windows) - existing CGO requirements from rtmididrv
**Project Type**: Single Go project (CLI application)
**Performance Goals**: Execute 50-command script in <5 seconds (excluding MIDI/AI execution time)
**Constraints**:
- Must not block playback goroutine during script execution
- Pre-validation must complete before any command execution
- Real-time progress feedback (command echo) required

**Scale/Scope**:
- Support 1000+ command scripts without memory issues
- Handle AI commands that may take 2-10 seconds each
- Three execution modes: interactive, piped-then-interactive (`cat file - | app`), piped-then-continue (`cat file | app`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ I. Incremental Development
**Status**: PASS
**Rationale**: Feature adds stdin detection and batch processing without disrupting existing interactive mode. Can be built incrementally: (1) stdin detection, (2) batch processor, (3) script file flag, (4) pre-validation. Each step independently testable.

### ✅ II. Collaborative Decision-Making
**Status**: PASS
**Rationale**: No architectural changes - extends existing command processing with new input source. Trade-offs documented (continue vs. exit behavior, validation timing). Developer approved approach during clarification session.

### ✅ III. Musical Intelligence with Creative Freedom
**Status**: PASS
**Rationale**: Batch mode preserves all musical functionality. AI commands execute inline in scripts, maintaining musical intelligence. Pattern loop continues playing after script execution (performance tool paradigm). No impact on MIDI timing or playback goroutine.

### ✅ IV. Pattern-Based Simplicity
**Status**: PASS
**Rationale**: Uses existing pattern loop synchronization. Script commands queue pattern changes at loop boundaries (existing mechanism). No changes to playback goroutine or mutex strategy. Batch execution happens in main goroutine.

### ✅ V. Learning-First Documentation
**Status**: PASS
**Rationale**: This plan documents stdin detection approach, Go idioms (bufio.Scanner, term.IsTerminal), and design rationale. Quickstart guide will provide implementation walkthrough. CLAUDE.md updated with batch mode in Phase listing.

### ✅ VI. AI-First Creativity
**Status**: PASS
**Rationale**: AI commands (`ai <prompt>`) work identically in batch and interactive modes. Scripts can combine manual commands for precision with AI for creativity. Enables workflows like "set initial pattern, ask AI to add tension, save result."

### 🟡 Technology Stack Compliance
**Status**: PASS with addition
**Addition**: `github.com/mattn/go-isatty` for terminal detection (cross-platform support including Cygwin/Git Bash)
**Rationale**: Required to distinguish piped input from terminal input. Chosen over `golang.org/x/term` for superior cross-platform support (handles Windows Git Bash and Cygwin terminal detection).

### ✅ Architecture Constraints
**Status**: PASS
**Rationale**: All permanent modules unchanged. Adds stdin processing in main.go only. Commands package already handles command execution. No new goroutines or state machines.

### ✅ Musical Constraints
**Status**: PASS
**Rationale**: Musical intelligence preserved - AI commands work in batch mode. Creative dissonance still supported. Default patterns, humanization, and swing unaffected.

## Project Structure

### Documentation (this feature)

```text
specs/002-batch-script-mode/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (complete)
├── research.md          # Phase 0 output (to be generated)
├── data-model.md        # Phase 1 output (to be generated)
├── quickstart.md        # Phase 1 output (to be generated)
├── contracts/           # Phase 1 output (to be generated)
│   ├── stdin-detection.md
│   └── command-execution.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Single Go project structure (existing)
main.go                  # Modified: add stdin detection, flag parsing, batch processor
commands/
├── commands.go          # Existing: command handler (minimal or no changes)
├── set.go               # Existing: unchanged
├── rest.go              # Existing: unchanged
├── clear.go             # Existing: unchanged
├── tempo.go             # Existing: unchanged
├── show.go              # Existing: unchanged
├── save.go              # Existing: may add overwrite warning
├── load.go              # Existing: unchanged
├── delete.go            # Existing: may add deletion warning
└── [other commands]     # Existing: unchanged

sequence/                # Existing: unchanged
playback/                # Existing: unchanged
midi/                    # Existing: unchanged
ai/                      # Existing: unchanged

patterns/                # Existing: pattern storage
test_basic.txt           # Existing: example script
test_cc.txt              # Existing: example script
```

**Structure Decision**: Single Go project with flat package structure (existing architecture). Batch mode logic added to main.go as helper functions (`isTerminal()`, `processBatchInput()`). Minimal changes to commands package for validation warnings. No new packages needed - this is input mode variation, not new domain logic.

## Complexity Tracking

> No constitutional violations requiring justification.

All gates pass cleanly. The feature extends existing command processing with new input sources (stdin/file) without architectural changes or new dependencies beyond standard Go extended libraries.

