# Contract: Stdin Detection

**Feature**: Batch/Script Mode for Command Execution
**Component**: Input source detection
**Date**: 2024-12-05

## Purpose

Define the contract for detecting stdin type (terminal vs pipe) and determining execution mode (interactive, batch-continue, batch-exit).

## Functions

### `isTerminal() bool`

**Purpose**: Detect if stdin is connected to a terminal (TTY)

**Implementation**:
```go
import "golang.org/x/term"

func isTerminal() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**Returns**:
- `true`: stdin is a TTY (terminal) → use interactive mode
- `false`: stdin is a pipe or file → use batch mode

**Edge cases**:
- Redirected file: `./app < file.txt` → returns `false`
- Piped input: `cat file | ./app` → returns `false`
- Normal terminal: `./app` → returns `true`

**Thread safety**: Read-only check, safe to call from main goroutine

---

### `detectExecutionMode() ExecutionMode`

**Purpose**: Determine which execution mode to use based on stdin type and command-line flags

**Implementation**:
```go
type ExecutionMode int

const (
    ModeInteractive ExecutionMode = iota
    ModeBatchContinue  // Not implemented in P1 - requires pipe-then-interactive detection
    ModeBatchExit
)

func detectExecutionMode(scriptFile string) ExecutionMode {
    // Priority 1: --script flag always means batch-exit
    if scriptFile != "" {
        return ModeBatchExit
    }

    // Priority 2: TTY means interactive
    if isTerminal() {
        return ModeInteractive
    }

    // Priority 3: Pipe or redirected file means batch-exit
    return ModeBatchExit
}
```

**Parameters**:
- `scriptFile string`: Value of --script flag (empty if not set)

**Returns**: ExecutionMode constant

**Decision tree**:
```
--script flag set?
  YES → ModeBatchExit
  NO  → stdin is TTY?
          YES → ModeInteractive
          NO  → ModeBatchExit
```

**Note**: `ModeBatchContinue` (P1 requirement: `cat file - | app`) is deferred to a later phase due to complexity of detecting when piped input is exhausted but stdin remains open.

**Thread safety**: Called once at startup, no concurrency concerns

---

## Execution Mode Contracts

### ModeInteractive

**Pre-conditions**:
- stdin is a TTY

**Behavior**:
- Use readline for input
- Show "> " prompt
- Process commands until "quit" or Ctrl+D
- Exit code 0 on normal exit

**Post-conditions**:
- Playback stopped cleanly
- MIDI resources closed
- No error if user quits normally

---

### ModeBatchExit

**Pre-conditions**:
- stdin is NOT a TTY (pipe or file), OR
- --script flag is set

**Behavior**:
- Read commands line-by-line from input
- Skip lines starting with `#` (comments)
- Skip empty lines
- Process each command via existing `Handler.ProcessCommand()`
- Track errors but continue processing
- Exit after all commands processed

**Post-conditions**:
- Exit code 0 if all commands succeeded
- Exit code 1 if any command failed
- All commands attempted (no early exit on error)
- Playback stopped cleanly
- MIDI resources closed

---

### ModeBatchContinue (FUTURE)

**Pre-conditions**:
- stdin is a pipe, AND
- stdin remains open after piped data exhausted

**Behavior**:
- Read piped commands until blocked
- Switch to readline for interactive input
- Continue until "quit" or Ctrl+D

**Status**: NOT IMPLEMENTED in initial version - requires advanced pipe state detection

---

## Error Handling

### Stdin detection errors

**Scenario**: `term.IsTerminal()` fails (extremely rare)

**Handling**: Assume non-TTY (default to batch mode)

```go
func isTerminal() bool {
    // Defensive: if check fails, assume NOT terminal
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**Rationale**: Safer to default to batch mode (exits after processing) than interactive mode (waits forever)

---

### Script file errors

**Scenario**: `--script file.txt` but file doesn't exist

**Handling**:
```go
f, err := os.Open(scriptFile)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", err)
    os.Exit(2)  // Exit code 2 = invalid arguments
}
```

**Exit code**: 2 (misuse of command)

---

## Testing Contracts

### Test: TTY Detection
```bash
# Should use interactive mode (shows prompt)
./interplay
```

**Expected**: Prompt appears, readline active

---

### Test: Piped Input
```bash
# Should process commands and exit
echo "show" | ./interplay
```

**Expected**: Pattern displayed, program exits

---

### Test: Script File
```bash
# Should process file and exit
./interplay --script test_basic.txt
```

**Expected**: Commands executed, program exits with code 0

---

### Test: Script File Not Found
```bash
# Should error immediately
./interplay --script missing.txt
```

**Expected**: Error message, exit code 2

---

### Test: Piped Commands with Errors
```bash
# Should process all commands despite errors
echo -e "set 1 C3\nset 999 C3\nshow" | ./interplay
```

**Expected**:
- First command succeeds
- Second command errors (invalid step)
- Third command succeeds (shows pattern)
- Exit code 1 (had errors)

---

## Dependencies

**Required packages**:
- `golang.org/x/term` (new dependency)
- `os` (stdlib)
- `flag` (stdlib)

**Add dependency**:
```bash
go get golang.org/x/term
```

---

## Backward Compatibility

**Existing behavior preserved**:
- Running `./interplay` with no args on a terminal → unchanged (interactive mode)
- All existing commands work identically in batch and interactive modes

**New behavior**:
- Piped input now works (previously exited immediately)
- `--script` flag is new (no existing behavior to break)

**Migration**: None required - purely additive feature
