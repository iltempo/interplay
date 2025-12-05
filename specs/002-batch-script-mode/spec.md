# Feature Specification: Batch/Script Mode for Command Execution

**Feature Branch**: `002-batch-script-mode`
**Created**: 2024-12-05
**Status**: Draft
**Input**: User description: "Add batch/script mode to pipe commands from files for testing and automation"

## Clarifications

### Session 2025-12-05

- Q: Should the system validate or restrict script content to prevent dangerous operations? → A: Basic validation - warn on potentially destructive operations (e.g., delete commands, save operations that overwrite existing patterns)
- Q: How should AI commands behave when executed in batch scripts? → A: Execute inline - `ai <prompt>` works normally in batch mode (note: may take several seconds per AI command)
- Q: When should batch execution stop vs. continue after errors? → A: Use runtime validation with graceful continuation - invalid commands are logged as errors but script execution continues with remaining commands (pragmatic approach, simpler than pre-execution validation)
- Q: What feedback should users receive during batch script execution? → A: Progress with command echo - show each command as it executes plus results/errors (can be refined in later iterations)
- Q: How should exit codes reflect partial script failures? → A: Scripts set up performance state then keep program running (playback loop continues); exit 0 only if script contains explicit `exit` command with no failures; exit 1 if any command failed; otherwise program continues running after script completes

### Implementation Notes

- **Validation Strategy**: Runtime validation chosen over pre-execution validation for simplicity. Commands are validated as they execute, errors logged, execution continues. This aligns with the graceful continuation requirement and avoids complex syntax analysis.
- **Terminal Detection**: `github.com/mattn/go-isatty` chosen over `golang.org/x/term` for superior cross-platform support (handles Windows Git Bash and Cygwin terminal detection reliably).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Execute Commands from File (Priority: P1)

Users can pipe commands from a text file to the application, have all commands execute sequentially, and then continue with interactive mode for further testing or refinement.

**Why this priority**: This is the core MVP functionality. Without this, users cannot leverage batch/script mode at all. It enables rapid testing iteration and automation workflows.

**Independent Test**: Create a file with 3-5 commands (e.g., `set 1 C3`, `show`, `save test`), pipe it to the application using `cat commands.txt - | ./interplay`, verify all commands execute in order, then verify the prompt remains active for additional manual commands.

**Acceptance Scenarios**:

1. **Given** a file `test.txt` containing valid commands, **When** user runs `cat test.txt - | ./interplay`, **Then** all commands execute in sequence and the application remains in interactive mode
2. **Given** a file with commands including comments (lines starting with `#`), **When** piped to the application, **Then** comments are ignored and only actual commands execute
3. **Given** an empty file, **When** piped to the application, **Then** application starts in normal interactive mode without errors

---

### User Story 2 - Non-Interactive Batch Execution (Priority: P2)

Users can execute a script of commands with the application continuing to run and play after script completion, useful for setting up performance states. Users can optionally include an `exit` command to terminate the application after script execution.

**Why this priority**: Enables automation workflows for performance setup (load pattern, configure settings, start playing) where the application continues running for live performance. Also supports testing scenarios where explicit exit is desired.

**Independent Test**: Create a test file without `exit` command, run `cat test.txt | ./interplay` (without the dash), verify all commands execute and application continues running with playback loop active. Create another test file with `exit` command at end, verify application exits cleanly with appropriate exit code (0 for success, 1 for errors).

**Acceptance Scenarios**:

1. **Given** a file with commands but no `exit` command, **When** user runs `cat test.txt | ./interplay` (without `-`), **Then** all commands execute and application continues running with playback loop active
2. **Given** a script that encounters an error, **When** executed in batch mode, **Then** application reports the error, continues execution, and exits with code 1 if script contains `exit` command
3. **Given** a script with save/load commands and `exit` at the end, **When** executed in batch mode, **Then** all file operations complete and application exits cleanly

---

### User Story 3 - Execute Script File with Flag (Priority: P3)

Users can run the application with a script file argument (e.g., `./interplay --script test.txt`) for a more explicit and discoverable way to run batch commands.

**Why this priority**: Improves usability and discoverability, but pipes already provide this functionality. This is a convenience feature for users less familiar with Unix pipes.

**Independent Test**: Run `./interplay --script test.txt`, verify same behavior as piping the file, check `./interplay --help` shows the script flag option.

**Acceptance Scenarios**:

1. **Given** a script file path, **When** user runs `./interplay --script test.txt`, **Then** commands execute as if piped from the file
2. **Given** a non-existent file path, **When** user runs `./interplay --script missing.txt`, **Then** application shows clear error message and exits
3. **Given** the `--help` flag, **When** user runs `./interplay --help`, **Then** script mode options are documented

---

### Edge Cases

- What happens when a script contains invalid commands? (Runtime validation with graceful continuation - invalid commands are logged as errors but script execution continues with remaining commands)
- How does the system handle very large script files (1000+ commands)? (Should process all commands without memory issues or timeouts)
- What happens when stdin is closed unexpectedly during batch execution? (Application should complete processing buffered commands and exit cleanly)
- How does the application handle command failures during execution (e.g., AI API errors, file not found)? (Log error clearly, skip the failing command, continue with remaining commands)
- What happens when batch mode tries to execute AI commands? (AI commands with `ai <prompt>` syntax execute normally inline, taking several seconds per command; batch execution waits for AI response before continuing)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept commands from stdin when data is piped to the application
- **FR-002**: System MUST process piped commands line-by-line in sequential order
- **FR-003**: System MUST ignore lines beginning with `#` as comments in piped input
- **FR-004**: System MUST continue to interactive mode after processing piped commands when stdin remains open (e.g., `cat file - | app`)
- **FR-005**: System MUST continue running with playback loop active after processing all piped commands when stdin closes, unless script contains explicit `exit` command (e.g., `cat file | app`)
- **FR-006**: System MUST validate commands during execution with graceful continuation (runtime validation) - invalid commands logged as errors, execution continues with remaining commands
- **FR-006a**: System MUST report command errors clearly during execution, log the error, and continue processing remaining commands in batch mode
- **FR-007**: System MUST exit with code 0 when script contains explicit `exit` command and no errors occurred; exit with code 1 if any command failed; otherwise continue running after script completes
- **FR-008**: System MUST support `--script <filepath>` flag to execute commands from a file
- **FR-009**: System MUST validate script file existence before attempting to read it
- **FR-010**: System MUST handle empty script files gracefully without errors
- **FR-011**: System MUST warn users before executing potentially destructive operations in batch mode (delete commands, save operations that would overwrite existing patterns)
- **FR-012**: System MUST support AI commands (`ai <prompt>`) in batch mode, executing them inline and waiting for completion before processing subsequent commands
- **FR-013**: System MUST echo each command to output as it executes in batch mode, providing real-time progress visibility
- **FR-014**: System MUST display command results and error messages immediately after each command completes
- **FR-015**: System MUST recognize `exit` command in scripts to explicitly terminate the application after script completion

### Key Entities

- **Script File**: A text file containing one command per line, with optional comment lines (starting with `#`) and optional `exit` command to terminate application
- **Command Buffer**: Internal queue of commands read from stdin or script file, processed sequentially
- **Execution Context**: Tracks whether application is in batch mode (processing piped input) or interactive mode
- **Performance State**: The musical configuration (pattern, tempo, settings) established by script that persists after script execution, with playback loop continuing

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can execute a 50-command script file in under 5 seconds (excluding command execution time for MIDI operations and AI API calls)
- **SC-002**: 100% of valid commands in a script file execute successfully in batch mode
- **SC-003**: Application correctly transitions from batch to interactive mode when using `cat file - | app` syntax, or continues with playback loop active when using `cat file | app` syntax without `exit` command
- **SC-004**: Users can automate performance setup workflows by creating reusable script files that establish musical state and continue playing
- **SC-005**: Application exits cleanly with appropriate exit codes (0 when `exit` command present with no failures, 1 when any command failed) when script includes explicit `exit` command
- **SC-006**: Users receive clear error messages for failed commands during batch execution, with all errors logged without stopping script execution
