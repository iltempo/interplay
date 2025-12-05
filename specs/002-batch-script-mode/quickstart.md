# Quickstart: Batch/Script Mode

**Feature**: Batch/Script Mode for Command Execution
**Audience**: Developers implementing this feature
**Date**: 2024-12-05

## Goal

Enable users to pipe commands from files or stdin to automate testing and pattern creation workflows.

## What's Being Built

Three ways to provide input to Interplay:

1. **Interactive mode** (existing): Type commands at a prompt
2. **Piped batch mode** (new): Pipe commands, program exits after processing
3. **Script file mode** (new): Execute commands from a file via `--script` flag

## Prerequisites

- Existing Interplay codebase (main.go, commands/, sequence/, etc.)
- Go 1.25.4 or later
- Understanding of stdin/stdout in Unix-like systems

## Implementation Phases

### Phase 0: Add Dependency ✅

```bash
go get golang.org/x/term
```

**Verify**:
```bash
go mod tidy
```

### Phase 1: Add Stdin Detection (main.go)

**Location**: Add helper function before `main()`

```go
import (
    "golang.org/x/term"
    "os"
)

// isTerminal returns true if stdin is a terminal (TTY)
func isTerminal() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

**Test**:
```go
// In main(), temporarily add:
fmt.Printf("Is terminal: %v\n", isTerminal())

// Run tests:
./interplay              # Should print: Is terminal: true
echo "" | ./interplay   # Should print: Is terminal: false
```

### Phase 2: Add Flag Parsing (main.go)

**Location**: Top of `main()`, before MIDI port listing

```go
import "flag"

func main() {
    // Parse command-line flags
    scriptFile := flag.String("script", "", "execute commands from file")
    flag.Parse()

    // Check if script file provided
    if *scriptFile != "" {
        fmt.Fprintf(os.Stderr, "TODO: Script mode not yet implemented\n")
        os.Exit(2)
    }

    // ... existing MIDI port code
}
```

**Test**:
```bash
./interplay --script test.txt
# Expected: "TODO: Script mode not yet implemented"
```

### Phase 3: Implement Batch Input Processor

**Location**: Create new function in main.go (before `main()`)

```go
import (
    "bufio"
    "io"
    "strings"
)

// processBatchInput reads and executes commands from reader
// Returns false if any command failed
func processBatchInput(reader io.Reader, handler *commands.Handler) bool {
    scanner := bufio.NewScanner(reader)
    hadErrors := false

    for scanner.Scan() {
        line := scanner.Text()
        line = strings.TrimSpace(line)

        // Skip empty lines
        if line == "" {
            continue
        }

        // Print comments (for user visibility)
        if strings.HasPrefix(line, "#") {
            fmt.Println(line)
            continue
        }

        // Process command
        if err := handler.ProcessCommand(line); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            hadErrors = true
        }
    }

    // Check for scanner errors
    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
        return false
    }

    return !hadErrors
}
```

**Test**: Can't test yet - need to integrate into main()

### Phase 4: Implement Script File Mode

**Location**: Replace TODO in main() flag handling

```go
if *scriptFile != "" {
    // Open script file
    f, err := os.Open(*scriptFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", err)
        os.Exit(2)
    }
    defer f.Close()

    // ... Set up MIDI, pattern, playback, handler (existing code)

    // Process script file
    success := processBatchInput(f, cmdHandler)

    // Exit with appropriate code
    if success {
        os.Exit(0)
    } else {
        os.Exit(1)
    }
}
```

**Test**:
```bash
# Create test file
echo "show" > test_simple.txt

# Run with script flag
./interplay --script test_simple.txt

# Expected: Pattern displayed, program exits
```

### Phase 5: Implement Piped Input Mode

**Location**: Modify main() after creating cmdHandler

**Current code**:
```go
// Read commands from stdin
err = cmdHandler.ReadLoop(os.Stdin)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error reading commands: %v\n", err)
    os.Exit(1)
}
```

**Replace with**:
```go
// Determine input mode
if isTerminal() {
    // Interactive mode (existing behavior)
    err = cmdHandler.ReadLoop(os.Stdin)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading commands: %v\n", err)
        os.Exit(1)
    }
} else {
    // Batch mode (piped input)
    success := processBatchInput(os.Stdin, cmdHandler)
    if !success {
        os.Exit(1)
    }
}
```

**Test**:
```bash
# Test piped input
echo "show" | ./interplay

# Expected: Pattern displayed, program exits

# Test with errors
echo "set 999 C3" | ./interplay
echo $?

# Expected: Error message, exit code 1
```

### Phase 6: Verify AI Commands Work in Batch Mode

**No code changes needed** - AI commands already work with new `ai <prompt>` syntax!

**Current implementation** (commands/commands.go handleAI):
- Old design: `ai` enters interactive mode (incompatible with batch)
- New design: `ai <prompt>` processes prompt directly (works in batch)

**Test**:
```bash
# Test AI in batch mode
cat > test_ai.txt << 'EOF'
set 1 C3
ai make it darker
show
EOF

./interplay --script test_ai.txt
# Expected: Pattern created, AI processes prompt, pattern displayed

# Test without API key
unset ANTHROPIC_API_KEY
echo "ai make it darker" | ./interplay
# Expected: Error message "AI not available (ANTHROPIC_API_KEY not set)"
```

**Key insight**: With the new `ai <prompt>` design, AI commands work identically in both interactive and batch modes - no special handling needed!

## Testing Checklist

### Smoke Tests

```bash
# 1. Interactive mode still works
./interplay
# Type "show", see pattern, Ctrl+D to exit

# 2. Piped input works
echo "show" | ./interplay

# 3. Script file works
./interplay --script test_basic.txt

# 4. Script file not found
./interplay --script missing.txt
# Expected: Error, exit code 2

# 5. Comments printed
echo "# comment" | ./interplay
# Expected: "# comment" printed to stdout
echo $?
# Expected: Exit code 0 (no errors)

# 6. Empty input
echo "" | ./interplay
# Expected: Exit code 0

# 7. Error handling
echo "set 999 C3" | ./interplay
echo $?
# Expected: Error message, exit code 1

# 8. Multiple commands
cat test_basic.txt | ./interplay
# Expected: All commands execute
```

### Integration Tests

```bash
# 1. Complex script
cat > complex.txt << 'EOF'
# Set up pattern
set 1 C3 vel:127
set 5 G2 vel:110

# Add humanization
humanize velocity 8
swing 50

# Show result
show

# Save
save batch-test
EOF

./interplay --script complex.txt
# Expected: All commands execute, pattern saved

# 2. Load and modify
cat > modify.txt << 'EOF'
load basic-bass
set 3 D#3
save bass-variation
EOF

./interplay --script modify.txt
# Expected: Pattern loaded, modified, saved
```

## Common Issues

### Issue 1: Readline EOF behavior

**Symptom**: Piped input causes immediate exit

**Fix**: Conditional use of readline (Phase 5 implementation)

### Issue 2: AI mode hangs in batch

**Symptom**: Script with `ai` command waits for input forever

**Fix**: Detect non-TTY in handleAI and return error (Phase 6)

### Issue 3: Comment handling

**Symptom**: Lines starting with # cause "unknown command" errors

**Fix**: Skip comments in processBatchInput (Phase 3)

## File Locations

**Modified files**:
- `main.go` - Add stdin detection, flag parsing, batch mode handling
- `commands/commands.go` - Update handleAI to check for TTY

**New functions**:
- `isTerminal()` in main.go
- `processBatchInput()` in main.go

**Test files** (already created):
- `test_basic.txt`
- `test_cc.txt`

## Dependencies Added

```go
import (
    "bufio"           // Line-by-line reading
    "flag"            // Command-line flag parsing
    "golang.org/x/term" // Terminal detection
    // ... existing imports
)
```

## Rollback Plan

If implementation causes issues:

1. **Remove flag parsing**:
   ```go
   // Comment out flag parsing in main()
   ```

2. **Revert ReadLoop call**:
   ```go
   // Restore original: err = cmdHandler.ReadLoop(os.Stdin)
   ```

3. **Remove new functions**:
   - Delete `isTerminal()`
   - Delete `processBatchInput()`

## Next Steps After Implementation

1. **Update README.md**: Document batch mode usage
2. **Add examples/**: Create example script files
3. **Consider P1 enhancement**: `cat file - | app` (pipe-then-interactive)
4. **User testing**: Share with early users for feedback

## Time Estimate

- Phase 1-2 (detection + flags): 15 minutes
- Phase 3 (batch processor): 15 minutes
- Phase 4 (script file): 10 minutes
- Phase 5 (piped input): 10 minutes
- Phase 6 (AI verification): 5 minutes
- Testing: 15 minutes

**Total**: ~70 minutes (1.5 hours)

**Complexity**: Low - well-defined changes, no new data structures
**Note**: AI commands work automatically with new `ai <prompt>` design - no special batch handling needed!
