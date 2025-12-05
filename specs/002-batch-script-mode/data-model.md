# Data Model: Batch/Script Mode

**Feature**: Batch/Script Mode for Command Execution
**Phase**: 1 (Design)
**Date**: 2024-12-05

## Overview

This feature does not introduce new data structures or persistent state. It modifies input handling to support three execution modes: interactive (terminal), batch-with-continuation (pipe + interactive), and batch-only (pipe + exit).

## Execution Modes

### Mode 1: Interactive (Current Default)

**Trigger**: Program started with TTY stdin (normal terminal session)

**Behavior**:
- Use readline for input (history, editing, prompt)
- Show "> " prompt
- Process commands one at a time
- Continue until "quit" command or Ctrl+D

**State**:
```go
type ExecutionMode int

const (
    ModeInteractive ExecutionMode = iota
    ModeBatchContinue
    ModeBatchExit
)
```

**No state changes required** - this is existing behavior

### Mode 2: Batch-with-Continuation

**Trigger**: `cat file - | ./interplay` (stdin is pipe, dash keeps stdin open)

**Behavior**:
1. **Batch phase**:
   - Read piped input line-by-line using bufio.Scanner
   - Process each command
   - Skip comments (lines starting with `#`)
   - Continue until scanner.Scan() blocks (no more buffered input)
2. **Transition phase**:
   - Detect stdin is still open
   - Switch to readline for interactive input
3. **Interactive phase**:
   - Identical to Mode 1

**State tracking**:
```go
// In main.go
mode := detectExecutionMode()

switch mode {
case ModeBatchContinue:
    // Process piped commands first
    processPipedInput(os.Stdin, cmdHandler)
    // Then switch to interactive
    startInteractiveMode(cmdHandler)
}
```

### Mode 3: Batch-Exit

**Trigger**: `cat file | ./interplay` OR `./interplay --script file.txt`

**Behavior**:
1. Read all commands from input source
2. Process each command
3. Track errors but continue processing
4. Exit with appropriate exit code:
   - `0` if all commands succeeded
   - `1` if any command failed

**State tracking**:
```go
// Track success/failure for exit code
var hadErrors bool

// Process commands
for scanner.Scan() {
    if err := cmdHandler.ProcessCommand(line); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        hadErrors = true
        // Continue with remaining commands
    }
}

// Exit with appropriate code
if hadErrors {
    os.Exit(1)
}
```

## Input Source Types

### Type 1: TTY (Terminal)

**Detection**:
```go
import "golang.org/x/term"

func isTTY() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**Characteristics**:
- Interactive user input
- Supports readline features (history, editing)
- Line-buffered input
- User sees prompt

### Type 2: Pipe

**Detection**:
```go
func isPipe() bool {
    stat, err := os.Stdin.Stat()
    if err != nil {
        return false
    }
    return (stat.Mode() & os.ModeNamedPipe) != 0
}
```

**Characteristics**:
- Data from another process (e.g., `cat file | app`)
- May close after EOF or remain open (if `cat file - |` syntax used)
- Buffered input - all available data read before blocking
- No prompt visible to user

### Type 3: File (via --script flag)

**Detection**: Flag value non-empty

```go
scriptFile := flag.String("script", "", "execute commands from file")
flag.Parse()

if *scriptFile != "" {
    // Open file explicitly
    f, err := os.Open(*scriptFile)
    // ... process like pipe
}
```

**Characteristics**:
- Explicit file path
- Always batch-exit mode (never continues to interactive)
- File may not exist (error on open)
- Clear user intent for batch execution

## Command Processing State

**No changes to command processing** - all three modes use existing `Handler.ProcessCommand()` method

**Existing state** (unchanged):
```go
// Handler in commands/commands.go
type Handler struct {
    pattern           *sequence.Pattern  // Shared pattern (next)
    verboseController VerboseController  // Playback engine
    aiClient          *ai.Client         // Optional AI client
}
```

**Thread safety**: Already handled via pattern mutex - no changes needed

## Error State

**Current behavior** (interactive mode):
- Command error → print to stderr
- Continue processing
- No exit on error

**New behavior** (batch modes):
- Command error → print to stderr
- Track error occurred (set `hadErrors = true`)
- Continue processing remaining commands
- Exit with code 1 if any errors occurred (batch-exit mode only)

**State variable**:
```go
var hadErrors bool  // Track if any command failed
```

## Configuration State

**Command-line flags**:
```go
var (
    scriptFile = flag.String("script", "", "execute commands from file")
    // Future: could add --no-interactive, --exit-on-error, etc.
)
```

**No persistent configuration** - behavior determined at startup only

## Transitions Between Modes

```
Start
  ↓
  ├─ TTY stdin? ──────────────────→ Interactive Mode
  │                                  (existing behavior)
  ├─ --script flag set? ────────────→ Batch-Exit Mode
  │                                  (process file, exit)
  └─ Pipe stdin?
       ├─ stdin remains open? ─────→ Batch-Continue Mode
       │                             (process pipe, then interactive)
       └─ stdin closes? ───────────→ Batch-Exit Mode
                                     (process pipe, exit)
```

**Detection sequence** (main.go):
```go
// 1. Check for --script flag (highest priority)
if *scriptFile != "" {
    return runBatchMode(scriptFile)
}

// 2. Check if stdin is TTY
if term.IsTerminal(int(os.Stdin.Fd())) {
    return runInteractiveMode()
}

// 3. Stdin is pipe - check if it remains open after initial read
return runPipedMode() // Handles both batch-continue and batch-exit
```

## Data Flow

### Interactive Mode (Existing)
```
User input (TTY)
  ↓
readline.Readline()
  ↓
Handler.ProcessCommand()
  ↓
Pattern.SetNote() / Pattern.SetCC() / etc.
  ↓
Pattern state (mutex-protected)
  ↓
Playback goroutine reads at loop boundary
```

### Batch Modes (New)
```
Input source (pipe or file)
  ↓
bufio.Scanner.Scan()
  ↓
Skip if comment (#) or empty
  ↓
Handler.ProcessCommand()  ← Same as interactive!
  ↓
Pattern.SetNote() / Pattern.SetCC() / etc.
  ↓
Pattern state (mutex-protected)
  ↓
Playback goroutine reads at loop boundary
```

**Key insight**: Command processing logic is IDENTICAL - only input source changes

## Summary: No New Data Structures

This feature is purely about **input handling**, not data modeling. The implementation:

1. **Reuses existing**: Handler, Pattern, command processing
2. **Adds mode detection**: Functions to detect TTY vs pipe vs file
3. **Adds mode switching**: Logic to transition from piped → interactive
4. **Adds error tracking**: Simple boolean flag for exit code

**Zero impact on**:
- Pattern state representation
- MIDI output
- AI integration
- Persistence format (JSON patterns)
- Playback loop behavior
