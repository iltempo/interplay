# Test Files for Batch/Script Mode

This directory contains test script files demonstrating various batch mode features.

## Test Files

### test_basic.txt
**Purpose**: Basic batch mode functionality
**Usage**: `./interplay --script test_basic.txt` or `cat test_basic.txt | ./interplay`

Demonstrates:
- Clear command
- Tempo setting
- Setting notes with velocity
- Gate length control
- Humanization
- Swing
- Pattern saving
- Transition to interactive mode (no exit command)

### test_cc.txt
**Purpose**: CC (Control Change) automation
**Usage**: `./interplay --script test_cc.txt` or `cat test_cc.txt | ./interplay`

Demonstrates:
- Per-step CC automation
- Filter sweep using CC#74 (filter cutoff)
- Resonance control using CC#71
- CC visualization with cc-show command
- Saving patterns with CC data

### test_exit.txt
**Purpose**: Exit command behavior
**Usage**: `./interplay --script test_exit.txt` or `cat test_exit.txt | ./interplay`

Demonstrates:
- Clean exit after script execution
- Exit code 0 on success
- Script mode that doesn't continue playing

### test_interactive.txt
**Purpose**: Interactive continuation mode
**Usage**: `cat test_interactive.txt - | ./interplay`
**Note**: The dash (`-`) after the filename is required!

Demonstrates:
- Piped input followed by interactive mode
- Setting up a foundation pattern via script
- Continuing with manual commands afterward
- Best for rapid iteration workflows

### test_errors.txt
**Purpose**: Error handling and graceful continuation
**Usage**: `./interplay --script test_errors.txt` or `cat test_errors.txt | ./interplay`

Demonstrates:
- Errors logged to stderr
- Script continues executing after errors
- Valid commands still execute
- Graceful degradation

Contains intentional errors:
- Invalid step number (999)
- Invalid note name (X99)

### test_warnings.txt
**Purpose**: Destructive operation warnings
**Usage**: `./interplay --script test_warnings.txt` or `cat test_warnings.txt | ./interplay`

Demonstrates:
- Warning when overwriting existing pattern
- Warning when deleting pattern
- Clean exit after operations

## Running Tests

### Quick Test
```bash
# Run basic test with continuous playback
./interplay --script test_basic.txt
# Press Ctrl+C to stop

# Run exit test (auto-exits)
./interplay --script test_exit.txt
```

### Validation Tests
```bash
# Test 1: Basic functionality
./interplay --script test_basic.txt &
PID=$!
sleep 3
kill $PID

# Test 2: Exit command
./interplay --script test_exit.txt
echo "Exit code: $?"  # Should be 0

# Test 3: Error handling
./interplay --script test_errors.txt 2>&1 | grep "Error:"

# Test 4: Warnings
./interplay --script test_warnings.txt 2>&1 | grep "Warning"

# Test 5: Interactive continuation
cat test_interactive.txt - | ./interplay
# Type commands at the prompt, then 'quit'
```

## Exit Behavior Reference

| Script Type | Exit Command | Behavior | Exit Code |
|-------------|--------------|----------|-----------|
| Piped input | No | Continues playing | 0 (on Ctrl+C) |
| Piped input | Yes, no errors | Exits cleanly | 0 |
| Piped input | Yes, had errors | Exits with errors | 1 |
| Script file | No | Interactive mode | User quits |
| Script file | Yes, no errors | Exits cleanly | 0 |
| Script file | Yes, had errors | Exits with errors | 1 |
| Script file | File not found | Error message | 2 |
| Interactive pipe (`-`) | N/A | Continues interactive | User quits |

## Script as Preset

Script files (using `--script` flag) work as **presets** that set up musical state and then transition to interactive mode. This lets you:
- Create reusable starting points for performances
- Load complex patterns quickly and continue editing
- Build libraries of musical ideas

To exit automatically after a script, include `exit` as the last command.

For the "performance tool paradigm" (setup pattern and let it play), use piped input instead: `cat script.txt | ./interplay`
