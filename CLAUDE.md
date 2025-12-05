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
Question: "Why 48 steps as the default pattern length?"
Answer: "48 steps = 3 bars of 16th notes. This gives enough resolution for triplets
         and complex rhythms while staying manageable. Use 'length' command to adjust."
Decision: Higher resolution by default, easily adjustable via command.
```

## Project Goal

Control a Waldorf Robot Mono Synthesizer via MIDI to experiment with musical ideas and eventually use in live performances. The system uses a pattern-based looping approach: a short pattern (1-4 bars) plays continuously, and modifications are queued and applied at the start of the next loop iteration. This allows real-time experimentation without requiring immediate synchronization or crossfading.

Initial focus is on exploration and playing around with the tool before considering live performance reliability.

## Architecture

### Core Modules (Permanent)

- `midi/` - MIDI connection and message sending (notes + CC)
- `sequence/` - Sequence state and manipulation logic (notes, velocity, gate, CC)
- `playback/` - Background loop that continuously plays the sequence
- `commands/` - CLI command parser (kept as foundation for AI execution)
- `ai/` - Natural language interpretation and command generation
- `main.go` - Orchestrates all components

**Note**: The `commands/` module was originally planned as temporary but is now permanent. It serves as the execution foundation that both direct user commands and AI-generated commands use. The AI doesn't replace commands—it generates them.

## Core Design Decisions

**Pattern-Based Looping:**
- Short patterns (1-4 bars) play continuously in a background goroutine
- Changes are queued and applied at the start of the next loop iteration (clean hard cuts, no crossfading needed)
- Loop boundary acts as the synchronization point - finish current pattern, start modified one
- Thread-safe pattern state shared between playback and command handler

**Initial Implementation (Phase 1):**
- Default 48 steps = 3 bars of 16th notes (higher rhythmic resolution for complex patterns)
- 80 BPM default tempo
- Each step: a note (e.g., C4, D#4) OR a rest/silence
- All notes same duration/gate initially
- Visual feedback: simple console output showing notes as they play

**Initial Command Set:**
- `set <step> <note>` - set step to a note (e.g., `set 1 C4`)
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

**Note Duration and Gate (How Long Notes Play):**
- **Duration**: How many steps the note spans (1 to pattern length). Default is 1 step.
  - `dur:1` = note plays for 1 step (16th note at default tempo)
  - `dur:4` = note plays for 4 steps (quarter note)
  - `dur:8` = note plays for 8 steps (half note)
- **Gate**: What percentage of the duration the note actually sounds (1-100%). Default is 90%.
  - Applied as: `soundingSteps = duration × (gate / 100)`, minimum 1 step
  - Example: `dur:4 gate:50` = note spans 4 steps but only sounds for 2 steps
- **Important**: For single-step notes (duration=1), gate has no practical effect since the minimum is 1 step. Gate only affects notes with duration > 1.
- To create staccato/short notes that span multiple steps: use low gate values (e.g., `dur:4 gate:25`)
- To create legato/connected notes: use high gate values (e.g., `dur:4 gate:100`)

**Timing Calculations:**
- At 80 BPM: 16th note = 187.5ms
- Step timing = (60,000ms / BPM) / 4 (for 16th notes)

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
- Implement 48-step pattern (3 bars at 80 BPM) for higher rhythmic resolution
- Background playback loop with pattern state
- Simple commands: set notes, add rests, change tempo
- Changes queued and applied at next loop iteration
- Console output showing notes as they play

**Default Starting Pattern:**
- All steps: rest (clean slate for building patterns)
- 48 steps provides room for complex rhythms, triplets, and multi-bar phrases

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

### Phase 4: MIDI CC Parameter Control (In Progress)

**Philosophy**: Build generic MIDI CC foundation first, then layer synth-specific intelligence through profiles in later phases.

**Phase 4a: Generic MIDI CC Control** (Current Focus)
- Send any MIDI CC (Control Change) message: CC number 0-127, value 0-127
- Per-step CC automation (like velocity/gate): different CC values per step
- Multiple CC parameters per step (filter + resonance + envelope on same step)
- Save/load CC data with patterns (JSON persistence)
- Visual feedback showing active CC values per step
- Works with any MIDI synthesizer without configuration

**Commands**:
```
cc <number> <value>              # Set global CC (e.g., cc 74 64 for filter)
cc-step <step> <number> <value>  # Set CC for specific step
cc-clear <step> <number>         # Remove CC automation from step
cc-show                          # Display all active CC automations
```

**Example Usage**:
```
> set 1 C3                       # Step 1: Note C3
> cc-step 1 74 127               # Step 1: Filter fully open (CC#74 = 127)
> cc-step 1 71 64                # Step 1: Medium resonance (CC#71 = 64)
> set 5 G2
> cc-step 5 74 20                # Step 5: Filter nearly closed (dark sound)
> save dark-sweep
```

**Phase 4b: Synth Profile System** (Future)
- Auto-detect connected synthesizer (MIDI device name matching)
- Load synth profiles from `profiles/` directory
- Profile format: JSON with CC mappings, parameter names, musical context
- Friendly parameter names: "filter" instead of "CC#74"
- AI uses profile to understand synth capabilities
- Ships with hand-crafted profiles: Waldorf Robot, generic subtractive synth

**Phase 4c: AI Sound Design Intelligence** (Future)
- Profile-aware natural language: "make it darker" → AI knows which CC
- Musical intent mapping: "add aggression" → increase resonance
- Parameter automation suggestions: "add movement" → LFO or filter sweep
- Creative dissonance: "make it harsher" → extreme resonance values
- Synth-specific techniques based on profile context

**Phase 4d: Interplay Profile Builder** (Separate Tool - Future)
- Standalone application for creating synth profiles
- AI-powered PDF extraction from MIDI implementation charts
- Interactive profile refinement and testing
- Exports JSON profiles for use in main Interplay app
- Separate repository: `interplay-profile-builder`

**Design Decisions**:
- **Generic first**: CC control works without profiles (plug-and-play principle)
- **Synth-agnostic**: Any MIDI device works, profiles enhance the experience
- **Separate tools**: Profile builder is optional, doesn't bloat main app
- **Loop boundary sync**: CC changes queue and apply at loop start (like notes)
- **JSON persistence**: CC data stored with patterns for reproducibility

**Technical Architecture for Phase 4a**:

**Data Model Extension**:
```go
// Existing Step structure (Phase 1-3)
type Step struct {
    Note     *int  // MIDI note number (nil = rest)
    Velocity int   // 0-127
    Gate     int   // Percentage 0-100
}

// Phase 4a: Add CC automation
type Step struct {
    Note       *int            // MIDI note number (nil = rest)
    Velocity   int             // 0-127
    Gate       int             // Percentage 0-100
    CCValues   map[int]int     // CC# → Value (e.g., 74 → 127 for filter)
}
```

**MIDI Message Flow**:
1. Pattern loop reaches step boundary
2. For each step with active notes or CC values:
   - Send Note On (if Note != nil)
   - Send CC messages (for each entry in CCValues map)
   - Queue Note Off based on gate length
3. CC messages sent alongside notes, same timing guarantees

**Command Implementation**:
- `cc <number> <value>` - Sets CC globally, applies to all future steps until changed
- `cc-step <step> <number> <value>` - Sets CC for specific step, stored in CCValues map
- Pattern persistence: Serialize CCValues map in JSON alongside notes/velocity/gate

**JSON Pattern Format (Phase 4a)**:
```json
{
  "name": "Dark Filter Sweep",
  "tempo": 80,
  "steps": [
    {
      "step": 1,
      "note": "C3",
      "velocity": 100,
      "gate": 90,
      "cc": {
        "74": 127,  // Filter cutoff fully open
        "71": 64    // Medium resonance
      }
    },
    {
      "step": 5,
      "note": "G2",
      "velocity": 80,
      "gate": 90,
      "cc": {
        "74": 20    // Filter cutoff nearly closed
      }
    }
  ]
}
```

**Future Extension Points** (for Phase 4b/4c):
- `profiles/` directory for synth-specific profiles
- Profile loading on synth detection
- AI integration point: translate "make it darker" → `cc-step` commands
- Profile format already designed, just not implemented yet

### Phase 5: Batch/Script Mode ✅ (Completed)
- ✅ Stdin detection for piped input vs terminal
- ✅ `--script` flag for explicit file execution
- ✅ Comment handling (lines starting with `#`)
- ✅ Error tracking with graceful continuation
- ✅ Command echo for progress feedback
- ✅ Exit command recognition for controlled termination
- ✅ Performance tool paradigm: scripts setup state, playback continues

**Batch Mode Features:**
- Pipe commands from files: `cat commands.txt | ./interplay`
- Interactive continuation: `cat commands.txt - | ./interplay`
- Script file flag: `./interplay --script commands.txt`
- Graceful error handling: log errors, continue execution
- Real-time progress: echo commands as they execute
- Exit control: explicit `exit` command or continue playing
- AI commands work inline: `ai make it darker` in scripts

**Usage Examples:**
```bash
# Pipe commands and continue with playback
echo "set 1 C4" | ./interplay

# Pipe commands then interact
cat setup.txt - | ./interplay

# Execute script file
./interplay --script performance-setup.txt

# Script with AI commands
echo -e "set 1 C3\nai add tension\nshow" | ./interplay
```

### Phase 6: Live Performance Features
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

## Active Technologies
- JSON files in `patterns/` directory (existing pattern persistence system) (001-midi-cc-control)

## Recent Changes
- 001-midi-cc-control: Added Go 1.25.4
