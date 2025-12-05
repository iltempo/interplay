# Interplay 🎹

🚧 **Work in Progress**

**Interplay is an AI-assisted creative tool for rapid music creation with MIDI synthesizers.** Connect any MIDI device—hardware synth, drum machine, or software instrument—and collaborate with AI to build musical loops through conversation and commands. No configuration, no music theory required, just ideas and creative exploration.

## What is Interplay?

Interplay transforms your creative process by combining AI musical intelligence with direct MIDI control. Talk to the AI in natural language about your musical ideas—"make it darker," "add tension," "create a bass line"—and it translates your intent into patterns that play immediately on your synthesizer. For precision work, drop into command mode for exact control over every note.

**How it works:**
- **AI-first creativity**: Describe what you want musically, and AI builds patterns that match your vision
- **Plug and play**: Connect your MIDI synth, select it from the list, and start creating
- **Musical intelligence**: AI understands harmony, rhythm, tension, and resolution—while embracing dissonance as a creative tool
- **Rapid iteration**: Build and modify 16-step patterns in real-time with instant feedback
- **Pattern-based looping**: Changes apply at loop boundaries—no timing anxiety, just creative flow
- **Hybrid control**: Switch seamlessly between AI conversation and direct commands (`set 1 C3`, `tempo 120`)

Interplay works with your synthesizer's full MIDI capabilities—notes, velocity, gate length, and synth-specific parameters. The AI helps you stay musically coherent while encouraging experimentation with dissonance, unconventional harmonies, and creative tension when that's what your music needs.

**Current Status:** Phase 3 Complete - AI integration with hybrid command/natural language control

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

## Usage

### Basic Commands

Control patterns with simple text commands:

```
> set 1 C3          # Set step 1 to note C3
> set 5 G3          # Set step 5 to note G3
> velocity 1 120    # Make step 1 louder
> gate 5 50         # Make step 5 staccato (50% gate)
> tempo 100         # Change to 100 BPM
> show              # Display current pattern
> <enter>           # Also displays current pattern
```

**Pattern Management:**
```
> save my_bassline  # Save current pattern
> load my_bassline  # Load a saved pattern
> list              # Show all saved patterns
> delete old_idea   # Delete a pattern
```

Full command list: type `help`

### AI Mode - Creative Collaboration

Interplay's AI mode is where the magic happens. Talk to the AI about your musical ideas in natural language, and it responds with patterns that match your creative vision.

**Setup** (one-time):
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

Get your API key from [Anthropic](https://www.anthropic.com/api) (separate from Claude Pro subscription).

**Enter AI mode:**
```
> ai
AI> create a dark bass line
```

**What you can do:**
- **Musical creativity**: "make it darker", "add tension", "add some movement", "create dissonance"
- **Music theory**: "what scale is this?", "transpose up a fifth", "add a passing note"
- **Direct commands**: `set 1 C2`, `tempo 120`, `show` - execute immediately without AI call
- **Pattern exploration**: "try something unexpected", "make it more minimal", "add complexity"
- **Press Enter** to show the current pattern

The AI understands musical concepts—harmony, rhythm, tension, resolution—and helps you explore both consonant and dissonant ideas:

```
AI> create a dark bass line
I'll create a brooding bass pattern in C minor with some rhythmic interest.
Executing 4 command(s):
  > set 1 C2
  > set 5 G2
  > set 9 C2
  > set 13 F2
Try it out!

AI> add some tension
Let me add a dissonant note to create tension before the resolution.
Executing 1 command(s):
  > set 7 Db2
This creates a half-step clash that builds anticipation!

AI> what scale is this in?
This is in C minor with a chromatic passing tone (Db). The dissonance adds tension!

AI> <enter>
[Shows current pattern with the tension-building dissonance]
```

**Alternative: Manual mode** - All commands work without an API key if you prefer direct control without AI assistance. Type `help` for the full command list.

## Learn More

See [CLAUDE.md](CLAUDE.md) for detailed development approach, architecture, and roadmap.