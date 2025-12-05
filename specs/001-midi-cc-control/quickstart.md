# Quickstart Guide: MIDI CC Parameter Control

**Feature**: 001-midi-cc-control | **For**: End Users | **Updated**: 2025-12-04

## What is MIDI CC Control?

MIDI CC (Control Change) messages let you control synthesizer parameters like filter cutoff, resonance, envelope attack, and more. This feature adds CC automation to Interplay, enabling you to create dynamic sound design alongside your note patterns.

**Key Concepts:**
- **CC Number**: Each parameter has a number (0-127). Example: CC#74 is usually filter cutoff
- **CC Value**: The parameter setting (0-127). Example: 127 = fully open filter
- **Global CC**: Quick experimentation - affects whole pattern, not saved
- **Per-Step CC**: Permanent automation - saved with pattern, varies per step

## Basic Workflow

### 1. Experiment with Global CC

Try different parameter values while your pattern plays:

```
> cc 74 127        # Open filter fully
> cc 74 20         # Close filter
> cc 71 80         # Increase resonance
```

**What happens:**
- Parameter changes take effect at next loop iteration
- Global CC values are transient (not saved with pattern)
- Perfect for live experimentation

### 2. Convert to Permanent Automation

Found a setting you like? Make it permanent:

```
> cc-apply 74      # Apply filter value to all steps with notes
> save my-pattern  # Now it's saved!
```

### 3. Create Dynamic Parameter Sweeps

Automate parameters per step for movement:

```
> set 1 C3
> set 5 G2
> set 9 C3
> set 13 F2
> cc-step 1 74 127    # Filter open on step 1
> cc-step 5 74 80     # Filter mid on step 5
> cc-step 9 74 40     # Filter closed on step 9
> cc-step 13 74 60    # Filter mid-open on step 13
> save filter-sweep
```

**Result**: Your pattern now has a filter sweep that plays every loop.

## Command Reference

### Global CC Commands

**Set global CC value** (transient, not saved):
```
cc <cc-number> <value>
```
Example: `cc 74 100` - Set filter cutoff to 100

**Apply global CC to all steps** (converts to permanent):
```
cc-apply <cc-number>
```
Example: `cc-apply 74` - Make current filter setting permanent for all note steps

### Per-Step CC Commands

**Set CC automation on specific step** (persistent, saved):
```
cc-step <step> <cc-number> <value>
```
Example: `cc-step 1 74 127` - Open filter fully on step 1

**Clear CC automation from step**:
```
cc-clear <step> <cc-number>
```
Example: `cc-clear 1 74` - Remove filter automation from step 1

**Clear all CC automation from step**:
```
cc-clear <step>
```
Example: `cc-clear 1` - Remove all CC automation from step 1

### Display Commands

**Show all CC automation**:
```
cc-show
```
Displays table of all active CC parameters per step

**Show pattern with CC indicators**:
```
show
```
Pattern display includes indicators for steps with CC automation

## Common CC Numbers

These CC numbers are standard across many synthesizers (check your synth's MIDI implementation chart):

| CC# | Parameter | Typical Use |
|-----|-----------|-------------|
| 1   | Modulation | Vibrato/modulation wheel |
| 7   | Volume | Overall volume |
| 10  | Pan | Stereo position |
| 71  | Resonance | Filter resonance/Q |
| 72  | Release Time | Envelope release |
| 73  | Attack Time | Envelope attack |
| 74  | Cutoff/Brightness | Filter cutoff frequency |
| 75  | Decay Time | Envelope decay |
| 91  | Reverb | Reverb send level |
| 93  | Chorus | Chorus send level |

**Note**: CC assignments vary by synthesizer. Check your synth's manual for its specific MIDI implementation chart.

## Example Sessions

### Example 1: Quick Filter Experiment

```
> load my-pattern
Pattern loaded: my-pattern (4 notes, 80 BPM)

> cc 74 50          # Try mid filter
> cc 74 127         # Try fully open - too bright!
> cc 74 80          # Perfect!
> cc-apply 74       # Make it permanent
> save my-pattern
Pattern saved: my-pattern.json
```

### Example 2: Create Filter Sweep Bassline

```
> clear
> length 16
> set 1 C2
> set 5 C2
> set 9 G1
> set 13 C2
> cc-step 1 74 127     # Start bright
> cc-step 5 74 90
> cc-step 9 74 50
> cc-step 13 74 70
> cc-step 1 71 40      # Add resonance to step 1
> save dark-bass
Pattern saved: dark-bass.json
```

### Example 3: Experiment Then Refine

```
> load groove-pattern
Pattern loaded: groove-pattern

> cc 74 100         # Experiment with filter
> cc 71 60          # And resonance
[Play and listen...]

> cc-apply 74       # Filter sounds good - apply it
> cc 71 80          # Try more resonance
> cc-apply 71       # Good! Apply that too
> save groove-pattern
Pattern saved: groove-pattern.json
```

### Example 4: Parameter-Only Steps (No Notes)

```
> set 1 C3
> set 9 C3
> cc-step 5 74 20      # Close filter mid-pattern (no note)
> cc-step 13 74 100    # Open filter (no note)
```

**Result**: Steps 5 and 13 only send CC messages, no notes. Creates parameter movement between note triggers.

## Tips & Best Practices

### Workflow Recommendations

1. **Start with global CC** - Experiment quickly without worrying about persistence
2. **Use cc-apply for consistent values** - Apply global CC when you want same value on all steps
3. **Use cc-step for automation** - Create dynamic sweeps and movement
4. **Save warning** - If you see a warning about unsaved global CC, use `cc-apply` before saving

### Creative Techniques

**Filter Sweeps:**
- Gradual filter opening creates tension and build
- Sudden filter close creates impact and drops
- Random filter values create unpredictability

**Resonance Control:**
- High resonance + low cutoff = dark, hollow sound
- Moderate resonance + sweeping cutoff = classic acid sound
- Low resonance = smooth, natural sound

**Envelope Shaping:**
- CC#73 (attack) + CC#72 (release) = control note shape per step
- Slow attack on some steps, fast on others = rhythmic variation
- Long release on last step = tail/decay at pattern end

**Multiple Parameters:**
- Combine filter + resonance for complex timbral changes
- Layer envelope + filter for dynamic articulation
- Use volume (CC#7) for accent patterns

### Common Patterns

**Classic Acid Bassline:**
```
cc-step 1 74 127
cc-step 2 74 100
cc-step 3 74 80
cc-step 4 74 60
cc-step 1 71 90    # High resonance on step 1
```

**Breathing Filter:**
```
cc-step 1 74 127   # Open
cc-step 5 74 40    # Closed
cc-step 9 74 127   # Open
cc-step 13 74 40   # Closed
```

**Volume Accents:**
```
cc-step 1 7 127    # Accent
cc-step 5 7 80     # Normal
cc-step 9 7 100    # Medium accent
cc-step 13 7 80    # Normal
```

## Troubleshooting

### Parameter Doesn't Change

**Problem**: Sent CC command but synth parameter didn't change

**Solutions:**
- Check your synth's MIDI implementation chart for correct CC number
- Verify synth is listening on MIDI channel 1
- Some synths require CC receive to be enabled in settings
- Try a different CC number (e.g., CC#1 modulation wheel - usually always works)

### Pattern Sounds Different After Loading

**Problem**: Saved pattern with global CC, but it sounds different after loading

**Cause**: Global CC values are transient (not saved)

**Solution**: Use `cc-apply <cc-number>` before saving to convert global CC to permanent per-step automation

### Warning When Saving

**Problem**: "Warning: Global CC values (CC#74) will not be saved"

**Meaning**: You have global CC values that won't persist

**Solution**: Run `cc-apply 74` (or whichever CC number) before saving, or ignore if you want it to be temporary

### Too Many CC Messages

**Problem**: Pattern feels sluggish or MIDI connection seems slow

**Cause**: Sending many CC messages per step can use MIDI bandwidth

**Solution**:
- Limit to 2-4 CC parameters per step
- Use CC automation only where needed (not every step)
- MIDI bandwidth is rarely an issue in practice

## Next Steps

### Learn More

- Read your synthesizer's MIDI implementation chart to discover available parameters
- Experiment with CC numbers to find interesting sonic territories
- Combine CC automation with humanization and swing for organic, evolving patterns

### Advanced Use Cases

- **AI Integration** (when enabled): Ask AI to create patterns with CC automation
  - "Create a dark bassline with a filter sweep"
  - "Add resonance automation to make it more aggressive"

- **Sound Design Experimentation**:
  - Set random CC values and discover new sounds
  - Create templates with different CC setups for different moods
  - Layer multiple CC parameters for complex timbral evolution

### Ask Questions

If you're unsure which CC number controls a parameter:
1. Check synth manual's MIDI implementation chart
2. Try common numbers (74=filter, 71=resonance) and listen
3. Use MIDI learn if your synth supports it (future Interplay feature)

## Pattern File Format

CC automation is stored in pattern JSON files like this:

```json
{
  "name": "filter-sweep",
  "tempo": 80,
  "length": 16,
  "steps": [
    {
      "step": 1,
      "note": "C2",
      "velocity": 100,
      "gate": 90,
      "cc": {
        "74": 127,
        "71": 64
      }
    },
    {
      "step": 5,
      "note": "G2",
      "velocity": 100,
      "gate": 90,
      "cc": {
        "74": 80
      }
    }
  ]
}
```

**Key Points:**
- `cc` field is optional (omitted if no CC automation)
- CC keys are CC numbers, values are parameter values
- Multiple CC parameters per step supported
- Fully backward compatible with patterns created before CC feature

You can manually edit these files if you prefer working with JSON directly.
