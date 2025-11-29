# Interplay 🎹

🚧 **Work in Progress**

Interplay is a conversational MIDI sequencer for live music creation and improvisation. It uses pattern-based looping and natural language control to make exploring musical ideas intuitive and immediate.

## What is Interplay?

Interplay lets you create and modify musical patterns through simple commands and (eventually) natural conversation. Patterns loop continuously, with changes queued and applied at loop boundaries—giving you real-time control without worrying about timing.

**Current Status:** Phase 1 - Basic pattern loop with MIDI playback working

## Installation

Check out the [releases page](https://github.com/iltempo/interplay/releases) for pre-built binaries, or install directly with Go:

```bash
go install github.com/iltempo/interplay@latest
```

Or build from source:

```bash
git clone https://github.com/iltempo/interplay.git
cd interplay
go build
./interplay
```

## Learn More

See [CLAUDE.md](CLAUDE.md) for detailed development approach, architecture, and roadmap.