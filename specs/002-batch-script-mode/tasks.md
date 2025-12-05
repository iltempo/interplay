# Tasks: Batch/Script Mode for Command Execution

**Input**: Design documents from `/specs/002-batch-script-mode/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md, contracts/

**Tests**: No test tasks included - feature specification does not explicitly request TDD approach. Manual testing scenarios provided in quickstart.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Single Go project structure at repository root
- Modified files: `main.go`, `commands/save.go`, `commands/delete.go`
- Test files: `test_basic.txt`, `test_cc.txt` (already exist)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add dependencies and prepare project for batch mode implementation

- [X] T001 Add `github.com/mattn/go-isatty` dependency via `go get github.com/mattn/go-isatty` and run `go mod tidy`
- [X] T002 [P] Update CLAUDE.md to document batch mode in Phase 4 listing with brief description

**Checkpoint**: Dependencies installed, documentation updated

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core stdin detection and flag parsing infrastructure that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Add `isTerminal()` helper function in main.go using `isatty.IsTerminal()` and `isatty.IsCygwinTerminal()` for cross-platform terminal detection
- [X] T004 Add flag parsing for `--script` flag in main.go at top of `main()` function before MIDI port listing
- [X] T005 Add `processBatchInput(reader io.Reader, handler *commands.Handler) (bool, bool)` function in main.go to process commands line-by-line using bufio.Scanner, returning (success, shouldExit)

**Checkpoint**: Foundation ready - stdin detection works, flag parsing in place, batch processor function exists

---

## Phase 3: User Story 1 - Execute Commands from File (Priority: P1) 🎯 MVP

**Goal**: Enable piping commands from text file to application with transition to interactive mode

**Independent Test**: Create file with 3-5 commands (`set 1 C3`, `show`, `save test`), run `cat commands.txt - | ./interplay`, verify all commands execute and prompt remains active for manual commands

### Implementation for User Story 1

- [X] T006 [US1] Modify main.go `main()` function to detect stdin mode using `isTerminal()` and route to batch vs interactive processing
- [X] T007 [US1] Implement piped input handling in main.go: when stdin is piped (not terminal), call `processBatchInput(os.Stdin, cmdHandler)`
- [X] T008 [US1] Add comment handling in `processBatchInput()`: skip lines starting with `#` and print them for visibility
- [X] T009 [US1] Add empty line handling in `processBatchInput()`: skip empty lines after `strings.TrimSpace()`
- [X] T010 [US1] Add command echo in `processBatchInput()`: print each command before execution for progress feedback
- [X] T011 [US1] Add error tracking in `processBatchInput()`: track errors with `hadErrors` bool but continue processing remaining commands
- [X] T012 [US1] Handle transition to interactive mode: after `processBatchInput()` completes for piped input, check if stdin still open and continue to existing `ReadLoop()`

**Checkpoint**: User Story 1 complete - can pipe commands from file with interactive continuation

**Manual Test Commands**:
```bash
# Test basic piping with interactive continuation
cat test_basic.txt - | ./interplay

# Test comment handling
echo "# This is a comment" | ./interplay

# Test empty file
echo "" | ./interplay
```

---

## Phase 4: User Story 2 - Non-Interactive Batch Execution (Priority: P2)

**Goal**: Enable script execution that continues running playback loop or exits based on `exit` command

**Independent Test**: Create test file without `exit`, run `cat test.txt | ./interplay`, verify commands execute and application continues running with playback active. Create test with `exit` at end, verify clean exit with correct exit code.

### Implementation for User Story 2

- [X] T013 [US2] Modify `processBatchInput()` to return success boolean based on `hadErrors` flag
- [X] T014 [US2] Update main.go piped input handling to check `processBatchInput()` return value and handle exit behavior
- [X] T015 [US2] Implement exit command recognition: detect `exit` command in `processBatchInput()` and set flag to terminate after processing
- [X] T016 [US2] Implement exit code logic in main.go: exit with code 0 if `exit` command present and no errors, code 1 if any errors occurred
- [X] T017 [US2] Modify default behavior to continue running with playback loop after batch processing completes (unless `exit` command present)
- [X] T018 [P] [US2] Add destructive operation warnings in commands/save.go: check if file exists before save in batch mode and warn user
- [X] T019 [P] [US2] Add destructive operation warnings in commands/delete.go: warn before deleting pattern files in batch mode

**Checkpoint**: User Story 2 complete - batch execution with configurable exit behavior and warnings

**Manual Test Commands**:
```bash
# Test batch mode that continues running
echo "set 1 C4" | ./interplay
# Should continue with playback loop

# Test batch mode with exit
echo -e "set 1 C4\nexit" | ./interplay
echo $?  # Should print 0

# Test error exit code
echo "invalid command" | ./interplay
echo $?  # Should print 1 (if exit command present)
```

---

## Phase 5: User Story 3 - Execute Script File with Flag (Priority: P3)

**Goal**: Enable explicit script file execution via `--script` flag for users unfamiliar with Unix pipes

**Independent Test**: Run `./interplay --script test.txt`, verify same behavior as piping, check `./interplay --help` shows script flag

### Implementation for User Story 3

- [X] T020 [US3] Implement script file mode in main.go: when `--script` flag is set, open file and validate it exists
- [X] T021 [US3] Add error handling for script file: print clear error message to stderr and exit with code 2 if file doesn't exist
- [X] T022 [US3] Route script file to `processBatchInput()`: pass file reader to batch processor function
- [X] T023 [US3] Handle script file exit behavior: always exit after processing (equivalent to `cat file | app` without interactive transition)
- [X] T024 [US3] Add help text documentation: update flag description to document `--script` flag usage
- [X] T025 [US3] Verify AI command compatibility: test that `ai <prompt>` commands work in script files (no code changes needed, verification only)

**Checkpoint**: User Story 3 complete - script file flag functional with help documentation

**Manual Test Commands**:
```bash
# Test script file mode
./interplay --script test_basic.txt

# Test non-existent file
./interplay --script missing.txt
# Expected: Error message, exit code 2

# Test help text
./interplay --help
# Should show --script flag documentation

# Test AI commands in script
cat > test_ai.txt << 'EOF'
set 1 C3
ai make it darker
show
EOF
./interplay --script test_ai.txt
```

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements and validation across all user stories

- [X] T026 [P] Add comprehensive example scripts in patterns/ directory: create example-batch-setup.txt with performance setup workflow
- [X] T027 [P] Update README.md: add batch mode section with usage examples for all three input modes
- [X] T028 Validate quickstart.md scenarios: run through all smoke tests and integration tests from quickstart.md
- [X] T029 Validate cross-platform compatibility: test on macOS terminal, Linux terminal, Windows Git Bash with piped input
- [X] T030 [P] Performance validation: create 50-command script and verify execution completes in under 5 seconds (excluding MIDI/AI time)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 (P1) → US2 (P2) → US3 (P3) must be sequential (each builds on previous)
  - US2 depends on US1 (extends piped input with exit behavior)
  - US3 depends on US1 (reuses batch processing logic)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational (Phase 2) - Foundation for all batch processing
- **User Story 2 (P2)**: Depends on US1 - Adds exit behavior and warnings to existing batch processing
- **User Story 3 (P3)**: Depends on US1 - Reuses batch processing with file flag instead of stdin

### Within Each User Story

- US1: Linear dependency (stdin detection → batch processing → comment/empty line handling → error tracking → interactive transition)
- US2: Mostly parallel (T018-T019 can run parallel, rest sequential)
- US3: Linear dependency (file handling → routing → exit behavior → help text → verification)

### Parallel Opportunities

- **Setup phase**: T001 and T002 can run in parallel
- **Foundational phase**: T003, T004, T005 must be sequential (T005 depends on understanding from T003-T004)
- **User Story 2**: T018 and T019 (warning additions) can run in parallel
- **Polish phase**: T026, T027, T030 can run in parallel

---

## Parallel Example: Setup Phase

```bash
# Launch setup tasks in parallel:
Task: "Add github.com/mattn/go-isatty dependency via go get"
Task: "Update CLAUDE.md to document batch mode"
```

## Parallel Example: User Story 2

```bash
# Launch warning additions in parallel:
Task: "Add destructive operation warnings in commands/save.go"
Task: "Add destructive operation warnings in commands/delete.go"
```

## Parallel Example: Polish Phase

```bash
# Launch polish tasks in parallel:
Task: "Add comprehensive example scripts in patterns/ directory"
Task: "Update README.md with batch mode documentation"
Task: "Performance validation with 50-command script"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (dependencies)
2. Complete Phase 2: Foundational (stdin detection, flag parsing, batch processor)
3. Complete Phase 3: User Story 1 (piped input with interactive continuation)
4. **STOP and VALIDATE**: Test US1 independently with manual test commands
5. This gives you working batch mode with interactive continuation - minimal viable feature

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **MVP delivered!**
3. Add User Story 2 → Test independently → Exit behavior + warnings added
4. Add User Story 3 → Test independently → Script file flag convenience added
5. Polish phase → Documentation, examples, validation
6. Each story adds value without breaking previous functionality

### Sequential Implementation (Recommended)

Since user stories build on each other:

1. Complete Setup (Phase 1)
2. Complete Foundational (Phase 2)
3. Complete User Story 1 (Phase 3) - Test and validate
4. Complete User Story 2 (Phase 4) - Test and validate
5. Complete User Story 3 (Phase 5) - Test and validate
6. Complete Polish (Phase 6)

This ensures each story is fully functional before adding the next layer of functionality.

---

## Notes

- **[P] tasks**: Different files or truly independent logic, can run in parallel
- **[Story] labels**: Map tasks to user stories for traceability
- **Sequential nature**: This feature has natural sequential dependencies (US1 → US2 → US3)
- **Manual testing**: Focus on manual testing per quickstart.md, no automated tests requested
- **File paths**: All tasks include specific file locations (main.go, commands/save.go, etc.)
- **Performance tool paradigm**: Remember this is for performance setup, not just batch processing
- **Exit behavior**: Key distinction - scripts setup state and continue playing unless explicit `exit` command

---

## Summary

**Total Tasks**: 30 tasks across 6 phases

**Task Count by User Story**:
- Setup: 2 tasks
- Foundational: 3 tasks (blocking)
- User Story 1 (P1): 7 tasks - **MVP scope**
- User Story 2 (P2): 7 tasks
- User Story 3 (P3): 6 tasks
- Polish: 5 tasks

**Parallel Opportunities**: 7 tasks marked [P] can run in parallel with others

**Independent Test Criteria**:
- US1: Pipe commands with interactive continuation
- US2: Batch execution with exit control and warnings
- US3: Script file flag with help documentation

**MVP Scope**: Complete through User Story 1 (Phases 1-3, tasks T001-T012) for minimum viable batch mode with interactive continuation

**Estimated Implementation Time**: ~70 minutes for MVP (from quickstart.md estimate)
