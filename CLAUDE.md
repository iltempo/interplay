# [CLAUDE.md](https://CLAUDE.md)

This file provides guidance to Claude Code ([claude.ai/code](https://claude.ai/code)) when working with code in this repository.

## Project Overview

This is a Go-based project for creating and improvising music with MIDI using AI. The project is in early stages with the foundational structure being established.

## Technology Stack

- **Language**: Go 1.25.4
- **Module**: `github.com/iltempo/interplay`

## Development Commands

Since the project is in early stages, standard Go commands apply:

```bash
# Build the project
go build ./...

# Run tests
go test ./...

# Run a specific test
go test -run TestName ./path/to/package

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## Development Approach

This project is being developed collaboratively with an emphasis on learning Go and understanding architectural decisions:

**Incremental Development:**
- Build one small piece at a time
- Each component is explained before implementation
- Review and iterate together before moving forward
- Make conscious decisions about architecture and design

**Learning Focus:**
- Go concepts explained as they're introduced (goroutines, channels, structs, interfaces, etc.)
- Discuss Go idioms and best practices in context
- Explore tradeoffs between different approaches
- Ask "why this way?" at any point

**Collaborative Decision-Making:**
- Discuss approach before implementing features
- Evaluate alternatives (e.g., "channel vs. mutex - what are the tradeoffs?")
- Developer makes final calls on architecture
- Understand implications of each choice

**Code Review Process:**
- Code proposed in small, digestible chunks
- Review, understand, and modify as needed
- Clarify anything unclear before proceeding
- Developer maintains control over what goes into the project

**Example Interaction:**
```
Question: "Why use an array instead of a slice for steps?"
Answer: "Arrays have fixed size, slices are dynamic. For our fixed 16 steps, 
         array is simpler. But if we want variable length later, we'd use a slice."
Decision: Start simple now, refactor when needed - or plan ahead if preferred.
```

## Project Goal

Control a Waldorf Robot Mono Synthesizer via MIDI to experiment with musical ideas and eventually use in live performances. The system uses a pattern-based looping approach: a short pattern (1-4 bars) plays continuously, and modifications are queued and applied at the start of the next loop iteration. This allows real-time experimentation without requiring immediate synchronization or crossfading.

Initial focus is on exploration and playing around with the tool before considering live performance reliability.

## Architecture

### Core Modules (Permanent)

- `midi/` - MIDI connection and message sending
- `sequence/` - Sequence state and note manipulation logic
- `playback/` - Background loop that continuously plays the sequence
- `main.go` - Orchestrates all components

### Temporary Modules (To Be Replaced)

- `commands/` - CLI command parser (will be replaced with AI in Phase 3)

This isolation allows easy replacement: when AI is integrated, delete `commands/` and add `ai/` that interprets natural language and calls the same underlying sequence manipulation functions.

## Core Design Decisions

**Pattern-Based Looping:**
- Short patterns (1-4 bars) play continuously in a background goroutine
- Changes are queued and applied at the start of the next loop iteration (clean hard cuts, no crossfading needed)
- Loop boundary acts as the synchronization point - finish current pattern, start modified one
- Thread-safe pattern state shared between playback and command handler

**Initial Implementation (Phase 1):**
- 1 bar = 16 steps (16th notes)
- 80 BPM default tempo
- Each step: a note (e.g., C4, D#4) OR a rest/silence
- All notes same duration/gate initially
- Visual feedback: simple console output showing notes as they play

**Initial Command Set:**
- `set <step> <note>` - set step 1-16 to a note (e.g., `set 1 C4`)
- `rest <step>` - make step silent
- `clear` - reset all to rests
- `tempo <bpm>` - change BPM
- `show` - display current pattern
- `quit` - exit program

## MIDI Architecture

**MIDI Library:**
- Using `gitlab.com/gomidi/midi/v2` (pure Go, cross-platform, well-maintained)
- Driver: `rtmididrv` (requires CGO, platform-specific builds)

**Distribution Considerations:**
- Current implementation requires CGO due to rtmididrv (RtMIDI C++ wrapper)
- Binary must be built per platform (macOS, Windows, Linux)
- Platform-specific system libraries required:
  - macOS: CoreMIDI, CoreAudio frameworks (always present)
  - Windows: Windows MIDI API (usually present)
  - Linux: ALSA (`libasound2`) or JACK (`libjack`)
- Binary size: ~12MB
- Alternative for static builds: `midicatdrv` (no CGO, requires midicat binary)
  - Trade-off: pure Go static binary vs. external midicat dependency
  - Decision: Keep rtmididrv for better performance and direct hardware access
  - Future: Consider build tags to support both drivers

**MIDI Output Configuration:**
- Channel: 1 (MIDI channel 0 in code)
- Velocity: Fixed at 100 (out of 127) initially
- Gate length: 90% of step duration (leaves small gap between notes)

**Timing Calculations:**
- At 80 BPM: 16th note = 187.5ms
- Step timing = (60,000ms / BPM) / 4 (for 16th notes)
- Note Off sent ~10ms before next step to avoid overlap

**Note Representation:**
- Input format: "C4", "D#5", "Bb3", etc.
- Internal storage: MIDI note number (C4 = 60, A4 = 69)
- Octave range: Typically C0-C8 (MIDI notes 12-108)

**Concurrency Model:**
- Main goroutine: handles user commands, updates "next pattern"
- Playback goroutine: continuously plays "current pattern"
- Pattern swap: at loop boundary (end of bar), current ← next
- Synchronization: mutex for safe pattern state access

## Development Phases

### Phase 1: Simple Pattern Loop (Current Focus)
- List available MIDI ports
- Send test notes to verify Waldorf receives them
- Implement 16-step pattern (1 bar at 80 BPM)
- Background playback loop with pattern state
- Simple commands: set notes, add rests, change tempo
- Changes queued and applied at next loop iteration
- Console output showing notes as they play

**Default Starting Pattern:**
- Steps 1, 5, 9, 13: C3 (root note on quarter beats)
- All other steps: rest
- Creates a simple bass pulse that's immediately musical and easy to build upon

### Phase 2: Musical Enhancements ✅ (Completed)
- ✅ Per-step duration/gate length
- ✅ Per-step velocity
- ✅ Pattern library (save/recall different patterns)
- ✅ Variable pattern lengths (via `length` command)
- ✅ Humanization (velocity, timing, gate randomization)
- ✅ Swing/Groove timing (adjustable swing percentage)

Note: Additional musical commands (transpose, randomize, etc.) can be handled via AI mode, so no need for dedicated commands.

**Pattern Persistence:**
- Format: JSON (human-readable, extensible, easy to share)
- File structure: `patterns/` directory with .json files
- Each file can contain one or multiple named patterns
- Metadata support: pattern name, tempo, tags, created date
- Commands:
    - `save <name>` - save current pattern to file
    - `load <name>` - load a saved pattern
    - `list` - show available saved patterns
    - `delete <name>` - remove a saved pattern

**JSON Pattern Format:**
```json
{
  "name": "Dark Bass Line",
  "tempo": 80,
  "steps": [
    {"step": 1, "note": "C3"},
    {"step": 5, "note": "C3"},
    {"step": 9, "note": "G2"},
    {"step": 13, "note": "C3"}
  ]
}
```

Future extensions can add: velocity, gate length, synth parameters per step

**Humanization:**
- Adds subtle random variations to make patterns feel more alive and organic
- Applied at playback time, so each loop iteration is slightly different
- Default settings (subtle):
    - Velocity: ±8 (range 0-64)
    - Timing: ±10ms (range 0-50ms)
    - Gate: ±5% (range 0-50%)
- Commands:
    - `humanize velocity <amount>` - set velocity randomization
    - `humanize timing <ms>` - set timing randomization
    - `humanize gate <amount>` - set gate randomization
    - `humanize` - show current settings
    - Set to 0 to disable
- Always enabled by default for more natural-sounding patterns

**Swing/Groove:**
- Delays every other 16th note (steps 2, 4, 6, 8, etc.) to create shuffle/groove feel
- Off by default - activate with `swing <percent>` command
- Common settings:
    - 0% = straight/even timing (default)
    - 50% = triplet swing (classic jazz/blues feel)
    - 66% = hard swing (very laid back)
- Commands:
    - `swing <percent>` - set swing amount (0-75%)
    - `swing` - show current setting
- Combines with humanization for even more organic feel
- Applied in playback engine, affects timing only

### Phase 3: AI Integration ✅ (Completed)
- ✅ AI natural language interpretation via Claude API
- ✅ AI translates user intent into pattern manipulation function calls
- ✅ Examples working: "make it darker" → transpose down, "add movement" → insert passing notes
- ✅ Hybrid approach implemented: direct commands + AI in unified mode
- ✅ Interactive AI session mode with conversation history
- ✅ Commands work directly in AI mode (no AI call needed)
- ⚠️ Commands package kept (not removed) - serves as foundation for AI execution
- 📝 AI as both preparation and live tool - works well for both use cases

**AI Mode Features:**
- `ai` command enters interactive AI session
- Direct commands execute immediately without calling AI
- Natural language sent to Claude for interpretation
- AI responds conversationally and can execute commands
- `[EXECUTE]...[/EXECUTE]` blocks contain commands to run
- Conversation history maintained across interactions
- `clear-chat` command resets conversation context
- Empty line (Enter) shows current pattern

### Phase 4: Advanced MIDI Features
- Synth parameter control via MIDI CC (filter, resonance, envelope, LFO)
- Integration with Waldorf MIDI implementation chart
- AI suggestions for synth parameters, not just notes

### Phase 5: Live Performance Features
- MIDI controller input (separate MIDI controller device)
    - Play notes on controller to add them to the pattern in real-time
    - Use knobs/faders to control synth parameters
    - AI-assisted parameter mapping: ask AI to randomly assign knobs to interesting parameters
    - Dynamic remapping during performance via dialogue with AI
- MIDI controller mapping for pattern control (pads for instant pattern recall, etc.)
- MIDI clock sync for playing with others
- Tap tempo
- Performance mode vs. rehearsal mode
- Keyboard shortcuts for critical operations
- Pattern transition options (immediate, next bar, next 4 bars)
