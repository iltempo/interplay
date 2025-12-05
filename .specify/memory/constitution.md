# Interplay Constitution

<!--
SYNC IMPACT REPORT
==================

Version Change: 1.0.0 → 1.1.0
Amendment Date: 2025-12-04
Reason: MINOR version bump - Added new AI-First Creativity principle and expanded Musical Constraints

Modified Principles:
- Principle III: "Real-Time Musical Reliability" → Enhanced to include musical intelligence and creative dissonance
- Technology Stack: AI Integration changed from "optional runtime dependency" to "core feature"

Added Sections:
- New Principle VI: AI-First Creativity - Elevates AI assistance from optional to foundational
- Musical Constraints: Expanded to include musical intelligence, dissonance, and creative tension

Removed Sections: None

Templates Requiring Updates:
- ✅ plan-template.md: Constitution Check section accommodates new principles
- ✅ spec-template.md: User scenarios support AI-first approach
- ✅ tasks-template.md: Phased approach compatible with AI-first development
- ✅ README.md: Already updated to reflect AI-first positioning

Follow-up TODOs: None

Rationale:
- README.md now positions Interplay as "AI-assisted creative tool" (AI-first)
- Musical intelligence with creative dissonance is now a core value
- AI mode described as "where the magic happens", not optional
- This constitutional amendment aligns governance with product positioning

Next Steps:
- Commit with: "docs: amend constitution to v1.1.0 (add AI-first creativity principle)"
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

### III. Musical Intelligence with Creative Freedom

Musical coherence and creative expression are equally valued. All features must:
- Maintain accurate MIDI timing (currently 16th-note precision)
- Preserve playback continuity during pattern changes
- Understand musical concepts: harmony, rhythm, tension, resolution
- Embrace dissonance, unconventional harmonies, and creative tension as valid musical tools
- Apply humanization/swing consistently and predictably
- Handle MIDI I/O without blocking the playback goroutine

**Rationale**: Interplay targets creative music-making, not just technically correct sequences. The AI must help users stay musically coherent while encouraging experimentation with dissonance when that serves the creative vision. Timing glitches destroy musical flow, but creative dissonance enhances it.

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

### VI. AI-First Creativity

AI assistance is a core feature, not an optional add-on. Design requirements:
- AI mode is the primary creative interface for pattern building
- Natural language interaction must translate musical intent to MIDI patterns
- Direct commands remain available for precision control and fallback
- AI must understand musical concepts and explain its creative decisions
- System must function without AI (degraded experience, not broken)

**Rationale**: Interplay is positioned as an "AI-assisted creative tool" where AI collaboration transforms the creative process. Users should be able to express musical ideas in natural language ("make it darker", "add tension") and receive musically intelligent responses. Manual commands serve precision needs and ensure the tool works without an API key.

## Development Constraints

### Technology Stack (MUST)

- **Language**: Go 1.25.4 (stdlib + minimal dependencies)
- **MIDI Library**: `gitlab.com/gomidi/midi/v2` with `rtmididrv` (CGO required)
- **AI Integration**: Anthropic Claude API via `anthropic-sdk-go` (core feature, graceful degradation without API key)
- **Distribution**: Platform-specific binaries (macOS/Windows/Linux) due to CGO
- **Testing**: Standard Go testing (`go test ./...`)

### Architecture Constraints (MUST)

- **Core modules are permanent**: `midi/`, `sequence/`, `playback/`, `main.go`, `ai/`
- **Pattern state**: Mutex-protected shared state between command handler and playback goroutine
- **Concurrency model**: Main goroutine (commands/AI) + playback goroutine (continuous loop)
- **Pattern swapping**: Current ← Next at loop boundary only
- **No external state**: Patterns stored as JSON in `patterns/` directory
- **AI integration**: Separate module that calls sequence manipulation functions

### Musical Constraints (MUST)

- **Musical intelligence**: AI must understand harmony, rhythm, tension, and resolution
- **Creative dissonance**: System embraces dissonant notes, chromatic passages, and unconventional harmonies as creative tools
- **Default pattern**: 16 steps (1 bar of 16th notes at configurable BPM)
- **Variable length**: Patterns support 1-64 steps via `length` command
- **Note format**: Human-readable (e.g., "C3", "D#4") converted to MIDI internally
- **Humanization**: Enabled by default with subtle settings (±8 velocity, ±10ms timing, ±5% gate)
- **Swing**: Off by default, configurable 0-75%
- **Musical explanations**: AI should explain creative choices (e.g., "This Db creates tension before resolution")

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

When using AI assistance (Claude Code or in-app AI mode):
- AI proposes, developer approves architectural changes
- Trade-offs must be explicit before implementation
- "Why this way?" questions are always welcome and must be answered
- Developer controls what goes into the codebase
- AI musical suggestions should include reasoning

## Governance

### Amendment Process

This constitution may be amended when:
1. A design decision establishes a new project-wide principle
2. A constraint proves limiting and alternatives are evaluated
3. Development phases reveal new architectural insights
4. Product positioning changes require principle updates

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
- Keep `CLAUDE.md` and `README.md` in sync with constitutional principles

**Version**: 1.1.0 | **Ratified**: 2025-12-04 | **Last Amended**: 2025-12-04
