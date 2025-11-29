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

## MIDI Setup

Interplay works with both **hardware MIDI devices** (synthesizers, drum machines, etc.) and **software instruments** (DAW plugins, virtual synths).

### Hardware MIDI Devices

Simply connect your MIDI device via USB or MIDI interface. Interplay will list all available ports when it starts—select your device.

### Software Instruments (Virtual MIDI)

To use Interplay with software synths in your DAW:

**macOS:**
1. Open **Audio MIDI Setup** (Applications → Utilities)
2. Go to **Window → Show MIDI Studio**
3. Double-click **IAC Driver**
4. Check **"Device is online"**
5. You now have virtual MIDI buses (e.g., "IAC Driver Bus 1")

**In your DAW:**
1. Create a MIDI track
2. Set the MIDI input to **IAC Driver Bus 1** (or whichever bus you created)
3. Add your software instrument/plugin to that track
4. Arm the track for recording (or enable MIDI monitoring)

**In Interplay:**
1. Run Interplay
2. Select **IAC Driver Bus 1** from the MIDI port list
3. Your patterns will now control the software instrument!

This lets you use Interplay with any VST/AU plugin, making it perfect for prototyping without hardware or exploring software synths.

## AI Features (Optional)

Interplay includes optional AI-powered features for natural language interaction. To enable:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

AI features require an [Anthropic API](https://www.anthropic.com/api) account (separate from Claude Pro subscription). Without an API key, all manual commands work normally.

## Learn More

See [CLAUDE.md](CLAUDE.md) for detailed development approach, architecture, and roadmap.