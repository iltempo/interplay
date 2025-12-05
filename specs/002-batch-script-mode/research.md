# Research: Batch/Script Mode Implementation

**Feature**: Batch/Script Mode for Command Execution
**Phase**: 0 (Technical Research)
**Date**: 2024-12-05

## Research Questions

### 1. How to detect if stdin is a terminal vs pipe in Go?

**Answer**: Use `os.Stdin.Stat()` to check file mode

```go
import (
    "os"
    "golang.org/x/term"
)

// Method 1: Using term.IsTerminal (recommended)
func isInteractive() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}

// Method 2: Using os.Stdin.Stat() (lower-level)
func isInteractive() bool {
    stat, _ := os.Stdin.Stat()
    return (stat.Mode() & os.ModeCharDevice) != 0
}
```

**Trade-offs**:
- `term.IsTerminal()`: Higher-level, more portable, requires `golang.org/x/term` package
- `os.Stdin.Stat()`: Lower-level, stdlib only, more explicit control

**Decision**: Use `term.IsTerminal()` for clarity and portability

**References**:
- https://pkg.go.dev/golang.org/x/term#IsTerminal
- https://stackoverflow.com/questions/43947363/detect-if-a-command-is-piped-to-stdin

### 2. Current readline library behavior with piped input

**Current Issue**: `github.com/chzyer/readline` immediately exits when it encounters EOF on piped input

**Analysis of current code** (commands/commands.go:840-863):
```go
func (h *Handler) ReadLoop(reader io.Reader) error {
    rl, err := readline.New("> ")  // Always creates readline instance
    if err != nil {
        return fmt.Errorf("failed to initialize readline: %w", err)
    }
    defer rl.Close()

    for {
        line, err := rl.Readline()
        if err != nil { // io.EOF or other error
            return nil  // EXIT on EOF!
        }
        // ... process command
    }
}
```

**Problem**: Readline reads directly from os.Stdin and exits on EOF, ignoring the `reader io.Reader` parameter completely

**Solutions**:
1. **Conditional readline**: Use readline only for interactive mode, use bufio.Scanner for piped input
2. **Custom readline input**: Configure readline with custom io.Reader (complex, not well-supported)
3. **Separate code paths**: Different functions for interactive vs batch mode

**Recommended Approach**: Solution 1 (conditional readline) - cleanest separation of concerns

### 3. Handling `cat file - | app` syntax (pipe then interactive)

**Explanation**: The dash (`-`) in `cat file -` tells cat to output the file contents then read from stdin

**Behavior**:
- Cat outputs all lines from `file` first
- Then waits for user input (the dash represents stdin)
- Both are piped together to the application

**Implementation Strategy**:
```
1. Detect stdin is NOT a terminal (it's a pipe)
2. Read all available piped input until it blocks (waiting for user input)
3. Once no more buffered input, check if stdin is STILL open
4. If stdin remains open, switch to interactive mode with readline
```

**Go Implementation**:
```go
// Use bufio.Scanner with non-blocking reads
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    line := scanner.Text()
    // Process command
}

// After scanner finishes, stdin might still be open
// Check if we should continue to interactive mode
if !stdinClosedDuringRead {
    // Switch to readline for interactive input
}
```

**Challenge**: Detecting when piped input is exhausted but stdin is still open requires non-trivial logic

### 4. Command-line flag parsing best practices in Go

**Standard library: `flag` package**

```go
import "flag"

var scriptFile = flag.String("script", "", "Execute commands from file")

func main() {
    flag.Parse()

    if *scriptFile != "" {
        // Open and read script file
        f, err := os.Open(*scriptFile)
        // ... handle file
    }
}
```

**Common patterns**:
- Use flag.String() for optional file paths
- flag.Parse() must be called before accessing flag values
- Remaining args available via flag.Args()

**Help output**: Automatically generated with `-h` or `--help` flag

**Alternative**: `spf13/cobra` for complex CLI apps (overkill for this use case)

**Decision**: Use stdlib `flag` package - simple and sufficient

### 5. Exit code conventions for CLI tools

**Standard exit codes**:
- `0`: Success
- `1`: General error (most common for application errors)
- `2`: Misuse of shell command (invalid arguments)
- `126`: Command found but not executable
- `127`: Command not found

**Best practice for this feature**:
- `0`: All commands executed successfully
- `1`: One or more commands failed (but execution continued)
- `2`: Invalid command-line arguments (e.g., script file not found)

**Implementation**:
```go
// Track errors during batch execution
var hadErrors bool

for /* each command */ {
    if err := processCommand(line); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        hadErrors = true
        // Continue processing remaining commands
    }
}

if hadErrors {
    os.Exit(1)
}
```

### 6. Comment handling in script files

**Requirement**: Print lines starting with `#` to provide context during script execution

**Implementation**:
```go
line := scanner.Text()
line = strings.TrimSpace(line)

// Skip empty lines
if line == "" {
    continue
}

// Print comments (provide context to user)
if strings.HasPrefix(line, "#") {
    fmt.Println(line)
    continue
}

// Process command
```

**Rationale**:
- Comments provide context when watching script execution
- Users can see what each section does while piping
- Helpful for debugging and understanding script flow

**Edge cases**:
- `# comment` - printed to stdout
- `  # indented comment` - printed (after TrimSpace)
- `set 1 C3 # inline comment` - process full line (inline comments NOT supported initially, can add later)

**Decision**: Start with line-level comments only (printed to stdout), consider inline comments if users request it

## Technical Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Stdin detection** | `term.IsTerminal()` | More portable, clearer intent |
| **Input handling** | Conditional: readline for TTY, bufio.Scanner for pipes | Clean separation, leverage existing readline for interactive |
| **`cat file - | app` support** | Detect exhausted pipe → switch to readline | Supports both batch and interactive in one session |
| **Flag parsing** | Stdlib `flag` package | Simple, sufficient, no added dependencies |
| **Exit codes** | 0=success, 1=had errors, 2=bad args | Standard CLI conventions |
| **Comments** | Line-level `#` printed to stdout (no inline) | Provides context during execution, simple parsing |

## Open Questions for Phase 1 Design

1. **AI mode in batch scripts**: ✅ RESOLVED
   - **Decision**: AI commands work seamlessly with `ai <prompt>` syntax
   - No mode switching needed - `ai` is just a command prefix
   - Conversation context maintained across multiple `ai` commands
   - Works identically in interactive and batch modes
   - Example: `ai make it darker` followed by `ai add more tension`

2. **Error handling strategy**: Continue on error or stop?
   - Current behavior (interactive): Show error, continue
   - **Decision**: Match interactive behavior - show error, continue with remaining commands

3. **Verbose output in batch mode**: Should `verbose` command work in scripts?
   - **Decision**: Yes - useful for debugging script execution

## Dependencies Required

**New dependency**:
- `golang.org/x/term` - Terminal detection

**Command to add**:
```bash
go get golang.org/x/term
```

**Existing dependencies** (already in use):
- `github.com/chzyer/readline` - Keep for interactive mode
- `os`, `bufio`, `flag` - Stdlib, already available

## Performance Considerations

**Memory**:
- bufio.Scanner default buffer: 64KB (sufficient for command scripts)
- No special handling needed for 1000+ commands

**Timing**:
- Command processing: <1ms per command (parse + pattern update)
- File I/O: Negligible for text files <1MB
- Target: <10ms overhead per command → easily achievable

**Concurrency**:
- No changes needed - commands already queue safely via mutex
- Pattern swap at loop boundary already thread-safe

## References

- Go stdlib `os` package: https://pkg.go.dev/os
- Go stdlib `bufio` package: https://pkg.go.dev/bufio
- Go stdlib `flag` package: https://pkg.go.dev/flag
- golang.org/x/term: https://pkg.go.dev/golang.org/x/term
- Exit code standards: https://tldp.org/LDP/abs/html/exitcodes.html
