# Interplay Constitution

<!--
SYNC IMPACT REPORT
==================

Version Change: Initial → 1.0.0
Ratification Date: 2025-12-04
Reason: Initial constitution ratification for Interplay project

New Principles Established:
1. Incremental Development - Build one small piece at a time with explanations
2. Collaborative Decision-Making - Developer maintains architectural control with AI assistance
3. Real-Time Musical Reliability - Priority on live performance and musical timing
4. Pattern-Based Simplicity - Queue changes for loop boundaries, avoid complex synchronization
5. Learning-First Documentation - Explain concepts as they're introduced

Templates Requiring Updates:
- ✅ plan-template.md: Constitution Check section present and ready
- ✅ spec-template.md: User scenarios and requirements align with incremental approach
- ✅ tasks-template.md: Phased approach supports incremental implementation

Follow-up TODOs: None

Next Steps:
- Commit with: "docs: establish Interplay constitution v1.0.0 (initial ratification)"
==================
-->

## Core Principles

### I. Incremental Development

Every feature is built one small piece at a time. Before implementation:
- Each component must be explained in context
- Go concepts (goroutines, channels, interfaces, etc.) introduced when needed
- Review and iterate together before moving forward
- Make conscious architectural decisions collaboratively

**Rationale**: This is a learning-focused project. Understanding precedes implementation. Incremental progress ensures both the AI assistant and developer maintain full comprehension of the system as it grows.

### II. Collaborative Decision-Making

The developer maintains final architectural control. Before significant changes:
- Discuss approach and evaluate alternatives with explicit trade-offs
- AI proposes solutions but MUST wait for developer approval on architecture
- Explain "why this way?" for all design decisions
- Developer decides scheduling and priority

**Rationale**: This project is as much about learning Go as building a tool. Architectural autonomy ensures the codebase reflects the developer's understanding and vision, not just AI suggestions.

### III. Real-Time Musical Reliability

Timing and musical feel are non-negotiable. All features must:
- Maintain accurate MIDI timing (currently 16th-note precision)
- Preserve playback continuity during pattern changes
- Apply humanization/swing consistently and predictably
- Handle MIDI I/O without blocking the playback goroutine

**Rationale**: Interplay targets live performance. Timing glitches destroy musical flow. Pattern changes queue at loop boundaries specifically to avoid synchronization complexity while maintaining reliability.

### IV. Pattern-Based Simplicity

The pattern loop is the synchronization primitive. Design choices:
- Changes queue and apply at loop iteration boundaries (start of bar)
- No crossfading, no immediate synchronization needed
- Thread-safe pattern state with mutex protection
- Background playback goroutine remains simple and focused

**Rationale**: Loop boundaries eliminate complex real-time synchronization. This architectural constraint trades flexibility for reliability and simplicity, which are paramount for live use.

### V. Learning-First Documentation

Documentation must teach, not just describe. Requirements:
- `CLAUDE.md` contains development approach, architectural decisions, and phase roadmap
- Code comments explain "why" for non-obvious Go idioms or trade-offs
- README serves user needs (quick start, MIDI setup, command reference)
- Design decisions are documented with alternatives considered

**Rationale**: This project has dual goals - functional tool AND learning experience. Documentation ensures knowledge transfers across sessions and provides context for future decisions.

## Development Constraints

### Technology Stack (MUST)

- **Language**: Go 1.25.4 (stdlib + minimal dependencies)
- **MIDI Library**: `gitlab.com/gomidi/midi/v2` with `rtmididrv` (CGO required)
- **AI Integration**: Anthropic Claude API via `anthropic-sdk-go` (optional runtime dependency)
- **Distribution**: Platform-specific binaries (macOS/Windows/Linux) due to CGO
- **Testing**: Standard Go testing (`go test ./...`)

### Architecture Constraints (MUST)

- **Core modules are permanent**: `midi/`, `sequence/`, `playback/`, `main.go`
- **Pattern state**: Mutex-protected shared state between command handler and playback goroutine
- **Concurrency model**: Main goroutine (commands) + playback goroutine (continuous loop)
- **Pattern swapping**: Current ← Next at loop boundary only
- **No external state**: Patterns stored as JSON in `patterns/` directory

### Musical Constraints (SHOULD)

- **Default pattern**: 16 steps (1 bar of 16th notes at configurable BPM)
- **Variable length**: Patterns support 1-64 steps via `length` command
- **Note format**: Human-readable (e.g., "C3", "D#4") converted to MIDI internally
- **Humanization**: Enabled by default with subtle settings (±8 velocity, ±10ms timing, ±5% gate)
- **Swing**: Off by default, configurable 0-75%

## Development Workflow

### Phase-Based Development (MUST)

Development proceeds in phases documented in `CLAUDE.md`:
- **Phase 1**: Simple pattern loop (complete)
- **Phase 2**: Musical enhancements (complete)
- **Phase 3**: AI integration (complete)
- **Phase 4**: Advanced MIDI features (planned)
- **Phase 5**: Live performance features (future)

Each phase must be complete before the next begins. Phases are reviewed and updated as understanding evolves.

### Code Review Process (MUST)

Before merging code:
- Propose code in small, digestible chunks
- Developer reviews, understands, and may modify
- Clarify anything unclear before proceeding
- Test incrementally (manual testing acceptable, automated tests optional)
- Commit messages follow conventional format: `feat:`, `fix:`, `docs:`, `refactor:`

### AI Assistance Protocol (MUST)

When using AI assistance (Claude Code or AI mode):
- AI proposes, developer approves architectural changes
- Trade-offs must be explicit before implementation
- "Why this way?" questions are always welcome and must be answered
- Developer controls what goes into the codebase

## Governance

### Amendment Process

This constitution may be amended when:
1. A design decision establishes a new project-wide principle
2. A constraint proves limiting and alternatives are evaluated
3. Development phases reveal new architectural insights

Amendments require:
- Documentation of reasoning and alternatives considered
- Update to this constitution with version bump
- Review of dependent templates (plan, spec, tasks)
- Git commit documenting the change

### Versioning

Constitution uses semantic versioning:
- **MAJOR**: Backward-incompatible governance changes (e.g., removing a principle)
- **MINOR**: New principle added or section materially expanded
- **PATCH**: Clarifications, wording improvements, non-semantic refinements

### Compliance

Development sessions should:
- Reference this constitution when making architectural decisions
- Justify any departures from established principles
- Propose amendments when constraints prove too limiting
- Keep `CLAUDE.md` in sync with architectural evolution

**Version**: 1.0.0 | **Ratified**: 2025-12-04 | **Last Amended**: 2025-12-04
