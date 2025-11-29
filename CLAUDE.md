# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based project for creating and improvising music with MIDI using AI. The project is in early stages with the foundational structure being established.

## Technology Stack

- **Language**: Go 1.25.4
- **Module**: `iltempo.de/midi-ai`

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

## Project Goal

Control a Waldorf Robot Mono Synthesizer via MIDI to experiment with musical ideas including melodies, chord sequences, and arpeggios. The system will start with a simple sequence and allow interactive, real-time evolution through text commands and eventually AI.

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

**Continuous Loop System:**
- Playback loop runs continuously in a background goroutine
- Starts with a simple default (e.g., C4 quarter notes at 120 BPM)
- User modifies sequence, tempo, and timing while it plays
- Changes take effect on the next loop iteration (clean, no glitches)
- Thread-safe sequence state shared between playback and command handler

**Initial Static Command Set:**

*Sequence modification:*
- `add <note>` - append note to sequence
- `insert <index> <note>` - insert note at position
- `remove <index>` - remove note at index
- `replace <index> <note>` - replace note at position
- `clear` - reset to empty/default
- `transpose <semitones>` - shift all notes

*Timing/tempo:*
- `tempo <bpm>` - set BPM
- `duration <beats>` - set note length

*Playback:*
- `pause` / `resume` - stop/start the loop
- `show` - display current sequence

## Development Phases

### Phase 1: MIDI Connectivity
- List available MIDI ports
- Send test notes to verify Waldorf receives them
- Confirm basic note on/off works

### Phase 2: Playback Loop + Static Commands
- Background playback loop with sequence state
- Simple CLI command parser (isolated in `commands/`)
- Real-time modifications applied on next loop iteration
- Adjustable tempo and note durations

### Phase 3: AI Integration
- Remove `commands/` package
- Add AI natural language interpretation
- AI translates user intent into sequence manipulation function calls
- Examples: "make it darker" → transpose down, "add movement" → insert passing notes

### Phase 4: Advanced MIDI Features
- Synth parameter control via MIDI CC (filter, resonance, envelope, LFO)
- Integration with Waldorf MIDI implementation chart
- AI suggestions for synth parameters, not just notes
