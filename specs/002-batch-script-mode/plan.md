# Implementation Plan: Batch/Script Mode for Command Execution

**Branch**: `002-batch-script-mode` | **Date**: 2024-12-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-batch-script-mode/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add support for executing commands from piped input or script files, enabling automated testing and batch command execution. Users can either pipe commands and continue with interactive mode (`cat file - | app`) or run batch scripts that exit automatically (`cat file | app`), and optionally use a `--script` flag for explicit file execution.

**Key benefit**: Makes AI workflows scriptable and repeatable. With the `ai <prompt>` design, users can save creative AI experiments as scripts and replay them, share AI workflows with others, and document their creative process naturally.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**:
- `github.com/chzyer/readline` (current input handling, needs modification)
- `os` package (stdin detection, file I/O)
- `bufio` (line-by-line reading)
- `flag` package (command-line argument parsing)

**Storage**: N/A (commands operate on existing pattern state)
**Testing**: Standard Go testing (`go test ./...`), manual testing with script files
**Target Platform**: macOS/Windows/Linux (same as existing)
**Project Type**: Single Go CLI application
**Performance Goals**: Process 1000+ commands without memory issues, minimal overhead (<10ms per command)
**Constraints**:
- Must not break existing interactive mode behavior
- Must detect stdin type (terminal vs pipe vs file)
- Must maintain thread-safety with playback goroutine
- Readline library currently causes immediate exit on EOF

**Scale/Scope**: Feature affects only main.go input handling and commands/commands.go ReadLoop method

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Incremental Development ✅
- **Status**: PASS
- **Assessment**: Small, focused feature with clear implementation phases (detect stdin → handle piped input → add flag support)
- **Justification**: Changes confined to input handling logic in main.go and commands package

### Principle II: Collaborative Decision-Making ✅
- **Status**: PASS
- **Assessment**: User explicitly requested this as a specification rather than quick fix, demonstrating thoughtful approach
- **Justification**: User tried workarounds first, then requested formal specification when behavior wasn't working

### Principle III: Musical Intelligence with Creative Freedom ✅
- **Status**: PASS
- **Assessment**: No impact on musical timing or playback continuity
- **Justification**: Batch mode processes commands that queue pattern changes at loop boundaries (existing mechanism)

### Principle IV: Pattern-Based Simplicity ✅
- **Status**: PASS
- **Assessment**: Uses existing pattern modification API, no changes to synchronization primitive
- **Justification**: Commands in batch mode use same queueing mechanism as interactive mode

### Principle V: Learning-First Documentation ✅
- **Status**: PASS
- **Assessment**: Will document stdin detection, pipe handling, and Go idioms for interactive vs batch mode
- **Justification**: Important learning topic: how CLI tools detect input source and adjust behavior

### Principle VI: AI-First Creativity ✅✅ (ENHANCED)
- **Status**: PASS WITH ENHANCEMENT
- **Assessment**: Batch mode ENABLES AI-first creativity by making AI workflows scriptable and repeatable
- **Justification**: New `ai <prompt>` design removes mode switching - AI commands work seamlessly in both interactive and batch modes
- **Impact**: This feature significantly enhances AI-first approach:
  - AI workflows can be saved and replayed
  - Creative experiments become reproducible
  - Users can share AI prompts as scripts
  - Natural language and native commands intermix naturally

**GATE RESULT**: ✅ PASS - All principles satisfied, no violations to justify

## Project Structure

### Documentation (this feature)

```text
specs/002-batch-script-mode/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── stdin-detection.md
│   └── command-execution.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Single project structure (existing)
main.go                  # MODIFY: Add stdin detection and flag parsing
commands/
└── commands.go          # MODIFY: ReadLoop method to handle piped input

# New test files (already created)
test_basic.txt
test_cc.txt
```

**Structure Decision**: Single Go CLI project. Changes isolated to main.go (input source detection and flag handling) and commands/commands.go (ReadLoop to support both piped and interactive input). No new packages required.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

N/A - No constitutional violations
