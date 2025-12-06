# Tasks: Model Comparison Framework

**Input**: Design documents from `/specs/003-model-comparison/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests not explicitly requested. Manual testing via quickstart.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is an existing Go CLI project. New files go in:
- `comparison/` - New module for comparison logic
- `ai/` - Extend existing AI client
- `commands/` - Extend existing command handler
- `main.go` - Add CLI flag handling

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new comparison module structure and model registry

- [ ] T001 Create comparison module directory at comparison/
- [ ] T002 [P] Create ModelConfig struct and AvailableModels registry in comparison/models.go
- [ ] T003 [P] Create Comparison and ModelResult structs in comparison/comparison.go
- [ ] T004 [P] Create Rating struct in comparison/rating.go
- [ ] T005 [P] Create BlindSession struct in comparison/blind.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend AI client with model switching - required by all user stories

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Add model field and SetModel/GetModel methods to ai.Client in ai/ai.go
- [ ] T007 Update GenerateCommands to use configurable model in ai/ai.go
- [ ] T008 Update Chat method to use configurable model in ai/ai.go
- [ ] T009 Update Session method to use configurable model in ai/ai.go
- [ ] T010 Add GetModelByID lookup function in comparison/models.go
- [ ] T011 Add --model flag parsing in main.go
- [ ] T012 Wire model flag to AI client initialization in main.go

**Checkpoint**: Model switching works via CLI flag. Foundation ready for user stories.

---

## Phase 3: User Story 1 - Run Comparison Test (Priority: P1) MVP

**Goal**: Execute same prompt against multiple models and save results

**Independent Test**: Run `compare "create a funky bass line"` and verify JSON saved to comparisons/

### Implementation for User Story 1

- [ ] T013 [US1] Implement generateComparisonID function in comparison/comparison.go
- [ ] T014 [US1] Implement RunComparison function that calls AI for each model in comparison/comparison.go
- [ ] T015 [US1] Implement executePromptForModel helper that captures commands and pattern in comparison/comparison.go
- [ ] T016 [US1] Implement SaveComparison function for JSON persistence in comparison/comparison.go
- [ ] T017 [US1] Add compare command handler in commands/commands.go
- [ ] T018 [US1] Implement progress feedback during comparison execution in commands/commands.go
- [ ] T019 [US1] Handle partial failures (continue with available models) in comparison/comparison.go
- [ ] T020 [US1] Handle parse errors (save raw response) in comparison/comparison.go
- [ ] T021 [US1] Create comparisons/ directory on first save in comparison/comparison.go

**Checkpoint**: User Story 1 complete. Can run comparisons and save results.

---

## Phase 4: User Story 2 - Review Saved Comparisons (Priority: P2)

**Goal**: List, view, load, and delete saved comparisons

**Independent Test**: Run `compare-list`, `compare-view <id>`, `compare-load <id> haiku`, `compare-delete <id>`

### Implementation for User Story 2

- [ ] T022 [P] [US2] Implement ListComparisons function in comparison/comparison.go
- [ ] T023 [P] [US2] Implement LoadComparison function in comparison/comparison.go
- [ ] T024 [P] [US2] Implement DeleteComparison function in comparison/comparison.go
- [ ] T025 [US2] Add compare-list command handler in commands/commands.go
- [ ] T026 [US2] Add compare-view command handler with formatted output in commands/commands.go
- [ ] T027 [US2] Add compare-load command handler (loads pattern to active state) in commands/commands.go
- [ ] T028 [US2] Add compare-delete command handler in commands/commands.go
- [ ] T029 [US2] Handle edge case: comparison not found in commands/commands.go
- [ ] T030 [US2] Handle edge case: model not in comparison in commands/commands.go

**Checkpoint**: User Story 2 complete. Can list, view, load, delete comparisons.

---

## Phase 5: User Story 3 - Select Model for Session (Priority: P3)

**Goal**: Switch AI model during session and list available models

**Independent Test**: Run `models`, `model sonnet`, then `ai make it darker` to verify model switched

### Implementation for User Story 3

- [ ] T031 [P] [US3] Implement GetActiveModel function in ai/ai.go
- [ ] T032 [US3] Add models command handler in commands/commands.go
- [ ] T033 [US3] Add model command handler in commands/commands.go
- [ ] T034 [US3] Show [active] marker for current model in models output in commands/commands.go
- [ ] T035 [US3] Validate model ID before switching in commands/commands.go

**Checkpoint**: User Story 3 complete. Can switch models at runtime.

---

## Phase 6: User Story 5 - Blind Evaluation Mode (Priority: P3)

**Goal**: Evaluate patterns without knowing which model produced them

**Independent Test**: Run `blind <id>`, load patterns A/B/C, rate them, `reveal` to see mapping

### Implementation for User Story 5

- [ ] T036 [P] [US5] Implement NewBlindSession with randomized label mapping in comparison/blind.go
- [ ] T037 [P] [US5] Implement GetPatternByLabel function in comparison/blind.go
- [ ] T038 [US5] Implement RateLabel function in comparison/blind.go
- [ ] T039 [US5] Implement Reveal function that returns label-to-model mapping in comparison/blind.go
- [ ] T040 [US5] Implement IsComplete check (all patterns rated) in comparison/blind.go
- [ ] T041 [US5] Add blind command handler that enters blind mode in commands/commands.go
- [ ] T042 [US5] Implement blind mode command loop (load, rate, reveal, exit) in commands/commands.go
- [ ] T043 [US5] Display anonymous pattern list on blind mode entry in commands/commands.go
- [ ] T044 [US5] Save ratings to comparison on reveal in commands/commands.go
- [ ] T045 [US5] Warn if only one model in comparison in commands/commands.go

**Checkpoint**: User Story 5 complete. Blind evaluation mode works.

---

## Phase 7: User Story 4 - Rate Comparison Results (Priority: P4)

**Goal**: Rate model outputs on musical criteria

**Independent Test**: Run `compare-rate <id> haiku rhythmic 4`, then `compare-view <id>` to see rating

### Implementation for User Story 4

- [ ] T046 [P] [US4] Implement AddRating function in comparison/rating.go
- [ ] T047 [P] [US4] Implement GetRating function in comparison/rating.go
- [ ] T048 [US4] Update SaveComparison to include ratings in comparison/comparison.go
- [ ] T049 [US4] Add compare-rate command handler in commands/commands.go
- [ ] T050 [US4] Validate criteria name (rhythmic, dynamics, genre, overall, all) in commands/commands.go
- [ ] T051 [US4] Validate score range 1-5 in commands/commands.go
- [ ] T052 [US4] Support "all" criteria to set all ratings at once in commands/commands.go
- [ ] T053 [US4] Display ratings in compare-view output in commands/commands.go

**Checkpoint**: User Story 4 complete. Can rate and view ratings.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, help text, and edge cases

- [ ] T054 [P] Add help text for all new commands in commands/commands.go
- [ ] T055 [P] Update README.md with comparison commands documentation
- [ ] T056 [P] Update CLAUDE.md with Phase 4 completion notes
- [ ] T057 Run quickstart.md validation (manual test all workflows)
- [ ] T058 Handle cancellation gracefully (Ctrl+C during comparison)
- [ ] T059 Add validation for empty prompt in compare command

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
- **Polish (Phase 8)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational - Depends on US1 for saved comparisons to exist
- **User Story 3 (P3)**: Can start after Foundational - Independent of other stories
- **User Story 5 (P3)**: Can start after Foundational - Depends on US1 for saved comparisons
- **User Story 4 (P4)**: Can start after Foundational - Depends on US1 for saved comparisons

### Within Each User Story

- Core logic in comparison/ module before command handlers
- Command handlers depend on comparison/ functions
- Each story complete before checkpoint

### Parallel Opportunities

- Setup phase: T002, T003, T004, T005 can run in parallel (different files)
- US2: T022, T023, T024 can run in parallel (different functions)
- US3: T031 can run parallel with T032-T035
- US5: T036, T037 can run in parallel
- US4: T046, T047 can run in parallel
- Polish: T054, T055, T056 can run in parallel

---

## Parallel Example: User Story 1

```bash
# All struct/function implementations can be written together:
Task: "Implement generateComparisonID function in comparison/comparison.go"
Task: "Implement RunComparison function in comparison/comparison.go"

# Then command handler after core logic:
Task: "Add compare command handler in commands/commands.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (5 tasks)
2. Complete Phase 2: Foundational (7 tasks)
3. Complete Phase 3: User Story 1 (9 tasks)
4. **STOP and VALIDATE**: Run `compare "test"` and verify JSON saved
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready (12 tasks)
2. Add User Story 1 → Test `compare` → MVP! (21 tasks total)
3. Add User Story 2 → Test `compare-list/view/load/delete` (30 tasks)
4. Add User Story 3 → Test `models`, `model` (35 tasks)
5. Add User Story 5 → Test `blind` (45 tasks)
6. Add User Story 4 → Test `compare-rate` (53 tasks)
7. Polish → Documentation complete (59 tasks)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- comparison/ module follows existing patterns/ module style
