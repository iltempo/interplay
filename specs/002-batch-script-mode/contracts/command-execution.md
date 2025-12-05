# Contract: Command Execution in Batch Mode

**Feature**: Batch/Script Mode for Command Execution
**Component**: Command processing and error handling
**Date**: 2024-12-05

## Purpose

Define how commands are read, parsed, and executed in batch mode, including comment handling, error tracking, and exit code behavior.

## Input Processing

### Function: `processBatchInput(reader io.Reader, handler *commands.Handler) bool`

**Purpose**: Read and execute commands from a non-interactive input source

**Parameters**:
- `reader io.Reader`: Input source (stdin pipe or opened file)
- `handler *commands.Handler`: Existing command handler

**Returns**:
- `bool`: `false` if any command failed, `true` if all succeeded

**Implementation contract**:
```go
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

        // Print comments (lines starting with #)
        if strings.HasPrefix(line, "#") {
            fmt.Println(line)  // Show comment to user
            continue
        }

        // Process command
        err := handler.ProcessCommand(line)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            hadErrors = true
            // CONTINUE processing remaining commands
        }
    }

    // Check for scanner errors (I/O errors, not command errors)
    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
        return false
    }

    return !hadErrors
}
```

**Guarantees**:
- All commands attempted (no early exit on error)
- Comments printed to stdout (for user visibility when piping)
- Empty lines skipped
- Each command processed independently
- Errors printed to stderr
- Scanner errors handled separately from command errors

---

## Comment Handling

### Rule: Line-level comments printed to stdout

**Supported**:
```bash
# This is a comment
  # This is also a comment (leading whitespace trimmed)

# Comments can explain sections
set 1 C3
set 5 G2
```

**Output when piping**:
```
# This is a comment
# This is also a comment (leading whitespace trimmed)
# Comments can explain sections
Set step 1 to C3 (velocity: 100)
Set step 5 to G2 (velocity: 100)
```

**Not supported** (processed as part of command):
```bash
set 1 C3 # This inline comment is NOT stripped
```

**Rationale**:
- Comments provide context when reviewing script output
- Users can see what each section does while watching execution
- Simple parser for MVP - inline comments can be added later if users request

**Implementation**:
```go
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "#") {
    fmt.Println(line)  // Print comment to stdout
    continue
}
// Otherwise process full line as-is
```

---

## Error Handling

### Command Errors

**Contract**: Continue on error, track for exit code

**Example scenario**:
```bash
set 1 C3      # Success
set 999 C3    # Error: step out of range
show          # Success - continues despite previous error
```

**Output**:
```
Set step 1 to C3 (velocity: 100)
Error: step number out of range: 999 (pattern length: 16)
[Pattern display...]
```

**Exit code**: 1 (because one command failed)

### I/O Errors

**Contract**: Stop processing, exit immediately

**Scenario**: File read error, disk full, etc.

**Handling**:
```go
if err := scanner.Err(); err != nil {
    fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
    os.Exit(1)
}
```

**Exit code**: 1

---

## Exit Codes

### Code 0: Success

**When**: All commands executed successfully

**Example**:
```bash
./interplay --script test_basic.txt
echo $?  # Output: 0
```

### Code 1: Command Error

**When**: One or more commands failed, OR I/O error occurred

**Example**:
```bash
echo "set 999 C3" | ./interplay
echo $?  # Output: 1
```

### Code 2: Argument Error

**When**: Invalid command-line arguments

**Example**:
```bash
./interplay --script nonexistent.txt
echo $?  # Output: 2
```

**Handling in main.go**:
```go
scriptFile := flag.String("script", "", "...")
flag.Parse()

if *scriptFile != "" {
    f, err := os.Open(*scriptFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(2)  // Argument error
    }
    // ... process file
}
```

---

## Integration with Existing Handler

### Contract: Reuse `Handler.ProcessCommand()`

**Existing signature** (commands/commands.go:41):
```go
func (h *Handler) ProcessCommand(cmdLine string) error
```

**Guarantees**:
- Parses command string
- Executes command via pattern API
- Returns error if command fails
- **Does NOT exit** on error (returns error instead)

**Batch mode usage**:
```go
err := handler.ProcessCommand(line)
if err != nil {
    // Batch mode: print error, continue
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    hadErrors = true
} else {
    // Command succeeded - normal output already printed by command
}
```

**Key insight**: ProcessCommand already has correct error handling for batch mode - we just need to track errors across multiple calls.

---

## Special Commands in Batch Mode

### `quit` command

**Interactive behavior**: Exit program

**Batch mode behavior**: Exit immediately (stop processing remaining commands)

**Implementation**:
```go
// In ProcessCommand, check for quit
if cmdLine == "quit" {
    return ErrQuit  // Special error type
}

// In processBatchInput
if err == commands.ErrQuit {
    return !hadErrors  // Exit cleanly, report error status
}
```

**Rationale**: Allow scripts to exit early if needed (e.g., conditional logic in future)

---

### `ai` command

**Current behavior**: Enter AI conversation mode (interactive only)

**New batch-friendly behavior**: `ai <prompt>` processes prompt and maintains conversation context

**Examples**:
```bash
# Interactive terminal
ai make it darker
# AI responds and may execute commands

# Batch script
ai make it darker
ai add chromatic movement
show
```

**Implementation** (in commands/commands.go handleAI):
```go
func (h *Handler) handleAI(parts []string) error {
    if len(parts) < 2 {
        return fmt.Errorf("usage: ai <prompt>")
    }

    // Join remaining parts as prompt
    prompt := strings.Join(parts[1:], " ")

    // Check if AI client available
    if h.aiClient == nil {
        return fmt.Errorf("AI not available (ANTHROPIC_API_KEY not set)")
    }

    // Process AI prompt with conversation context
    return h.processAIPrompt(prompt)
}
```

**Conversation context**:
- Multiple `ai` commands share conversation history
- Use `clear-chat` to reset context
- Context maintained across both interactive and batch modes

**Rationale**:
- No mode switching needed
- Works seamlessly in batch scripts
- Maintains conversational character
- Graceful degradation without API key

---

### `load` and `save` commands

**Batch mode behavior**: Work normally

**Example valid script**:
```bash
# Load base pattern
load basic-bass

# Modify it
set 3 D#3
set 7 F3

# Save as new pattern
save bass-variation-1
```

**Guarantee**: File I/O works identically in batch and interactive modes

---

## Output Behavior

### Standard output

**Contract**: Commands print results to stdout (same as interactive mode)

**Example**:
```bash
echo "show" | ./interplay
```

**Output**: Pattern visualization (same as typing `show` interactively)

### Error output

**Contract**: Errors print to stderr

**Example**:
```bash
echo "set 999 C3" | ./interplay 2>/dev/null
```

**Output**: (nothing - error redirected)

**Rationale**: Allows users to filter errors vs. normal output

---

## Concurrency

### Contract: Single-threaded command processing

**Guarantees**:
- Commands processed sequentially (one at a time)
- Pattern updates queue via existing mutex
- Playback goroutine continues independently

**Thread safety**:
- Batch processing runs in main goroutine
- `Handler.ProcessCommand()` already thread-safe (uses pattern mutex)
- No new concurrency concerns

---

## Performance

### Contract: Minimal overhead per command

**Target**: <10ms overhead per command (excluding actual command execution)

**Expected performance**:
- bufio.Scanner: ~0.1ms per line
- String parsing: ~0.1ms
- Pattern mutex: <0.1ms
- **Total overhead**: <1ms per command

**Validation**: 1000-command script should complete in ~1 second + actual MIDI operation time

---

## Testing Contracts

### Test: Empty File
```bash
touch empty.txt
./interplay --script empty.txt
echo $?  # Expected: 0
```

### Test: Comments Only
```bash
echo "# Just a comment" | ./interplay
# Expected output: "# Just a comment"
echo $?  # Expected: 0
```

### Test: Mixed Valid/Invalid
```bash
cat > test.txt << 'EOF'
set 1 C3
set 999 C3
set 5 G2
EOF

./interplay --script test.txt
echo $?  # Expected: 1 (had error)
```

### Test: Multi-line Command (NOT SUPPORTED)
```bash
# This will NOT work - each line is separate command
echo "set 1" | ./interplay  # Error: incomplete command
```

**Note**: Commands must be complete on a single line. No multi-line support.

---

## Backward Compatibility

**Existing command behavior**: Unchanged

**New behavior**: Batch mode uses same command processing, just different input source

**Breaking changes**: None - purely additive feature
