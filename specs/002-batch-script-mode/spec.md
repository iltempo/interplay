# Feature Specification: Batch/Script Mode for Command Execution

**Feature Branch**: `002-batch-script-mode`
**Created**: 2024-12-05
**Status**: Draft
**Input**: User description: "Add batch/script mode to pipe commands from files for testing and automation"

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

Users can execute a script of commands and have the application exit automatically after completion, useful for automated testing and CI/CD pipelines.

**Why this priority**: Enables automation workflows where no human interaction is needed. Critical for testing and integration scenarios but not required for basic manual testing workflow.

**Independent Test**: Create a test file, run `cat test.txt | ./interplay` (without the dash), verify all commands execute, verify application exits cleanly with appropriate exit code (0 for success, non-zero for errors).

**Acceptance Scenarios**:

1. **Given** a file with commands, **When** user runs `cat test.txt | ./interplay` (without `-`), **Then** all commands execute and application exits automatically
2. **Given** a script that encounters an error, **When** executed in batch mode, **Then** application reports the error and exits with non-zero exit code
3. **Given** a script with save/load commands, **When** executed in batch mode, **Then** all file operations complete before application exits

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

- What happens when a piped file contains an invalid command midway through execution? (Application should report error, skip the invalid command, and continue with remaining commands)
- How does the system handle very large script files (1000+ commands)? (Should process all commands without memory issues or timeouts)
- What happens when stdin is closed unexpectedly during batch execution? (Application should complete processing buffered commands and exit cleanly)
- How does the application handle mixed valid/invalid commands in a script? (Process valid ones, report invalid ones, continue execution)
- What happens when batch mode tries to enter AI mode (which also expects user input)? (Should either skip/warn, or buffer AI responses if provided in script)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept commands from stdin when data is piped to the application
- **FR-002**: System MUST process piped commands line-by-line in sequential order
- **FR-003**: System MUST ignore lines beginning with `#` as comments in piped input
- **FR-004**: System MUST continue to interactive mode after processing piped commands when stdin remains open (e.g., `cat file - | app`)
- **FR-005**: System MUST exit cleanly after processing all piped commands when stdin closes (e.g., `cat file | app`)
- **FR-006**: System MUST report command errors clearly but continue processing remaining commands in batch mode
- **FR-007**: System MUST exit with code 0 on successful batch execution and non-zero on errors
- **FR-008**: System MUST support `--script <filepath>` flag to execute commands from a file
- **FR-009**: System MUST validate script file existence before attempting to read it
- **FR-010**: System MUST handle empty script files gracefully without errors

### Key Entities

- **Script File**: A text file containing one command per line, with optional comment lines (starting with `#`)
- **Command Buffer**: Internal queue of commands read from stdin or script file, processed sequentially
- **Execution Context**: Tracks whether application is in batch mode (processing piped input) or interactive mode

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can execute a 50-command script file in under 5 seconds (excluding command execution time for MIDI operations)
- **SC-002**: 100% of valid commands in a script file execute successfully in batch mode
- **SC-003**: Application correctly transitions from batch to interactive mode when using `cat file - | app` syntax
- **SC-004**: Users can automate testing workflows by creating reusable script files that execute without manual intervention
- **SC-005**: Application exits cleanly with appropriate exit codes (0 for success, 1+ for errors) in batch mode
