# Tasks: MIDI CC Parameter Control

**Input**: Design documents from `/specs/001-midi-cc-control/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are NOT explicitly requested in the specification. Manual testing with MIDI hardware will be performed.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

This is a single Go project with the following structure:
- Core modules: `midi/`, `sequence/`, `playback/`, `commands/`, `ai/`
- Pattern storage: `patterns/` (JSON files)
- All paths are relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Review existing codebase and prepare for CC feature implementation

- [ ] T001 Review existing sequence/step.go structure to understand current Step definition
- [ ] T002 Review existing sequence/sequence.go to understand pattern state management and mutex usage
- [ ] T003 [P] Review existing midi/ package to understand MIDI message sending patterns
- [ ] T004 [P] Review existing commands/ package to understand command parser framework
- [ ] T005 [P] Review existing playback/ package to understand step playback loop

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data structures that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Extend Step struct with CCValues map[int]int field in sequence/step.go
- [ ] T007 Update Step JSON marshaling to include cc field with omitempty tag in sequence/step.go
- [ ] T008 Update Step JSON unmarshaling to handle optional cc field in sequence/step.go
- [ ] T009 Extend Sequence struct with globalCC map[int]int field in sequence/sequence.go
- [ ] T010 Implement CC validation helper function (0-127 range check) in sequence/validation.go
- [ ] T011 Add SendCC(channel, ccNumber, value uint8) method to midi/ package
- [ ] T012 Update playback loop to send CC messages before Note On in playback/playback.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Send Global CC Messages (Priority: P1) 🎯 MVP

**Goal**: Users can send global CC commands (e.g., `cc 74 127`) that affect the entire pattern and take effect at the next loop iteration

**Independent Test**: Connect MIDI synth, run pattern, send `cc 74 127`, verify synth parameter changes (filter opens). Global CC is transient (not saved).

### Implementation for User Story 1

- [ ] T013 [US1] Implement SetGlobalCC(ccNumber, value int) method with mutex protection in sequence/sequence.go
- [ ] T014 [US1] Implement GetGlobalCC(ccNumber int) method in sequence/sequence.go
- [ ] T015 [US1] Add cc command parser in commands/cc.go (parse: `cc <cc-number> <value>`)
- [ ] T016 [US1] Implement cc command handler that calls SetGlobalCC with validation in commands/cc.go
- [ ] T017 [US1] Update playback loop to send global CC messages at loop start in playback/playback.go
- [ ] T018 [US1] Add error handling for invalid CC numbers/values (display error without disrupting playback) in commands/cc.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently - users can send global CC and hear synth respond

---

## Phase 4: User Story 2 - Per-Step CC Automation (Priority: P2)

**Goal**: Users can create dynamic parameter automation where different steps have different CC values (e.g., filter sweeps)

**Independent Test**: Set different CC values on steps 1 and 5 (`cc-step 1 74 127`, `cc-step 5 74 20`), play pattern, hear filter sweep between steps

### Implementation for User Story 2

- [ ] T019 [US2] Implement SetStepCC(step, ccNumber, value int) method with mutex protection in sequence/sequence.go
- [ ] T020 [US2] Implement GetStepCC(step, ccNumber int) method in sequence/sequence.go
- [ ] T021 [US2] Add cc-step command parser in commands/cc_step.go (parse: `cc-step <step> <cc-number> <value>`)
- [ ] T022 [US2] Implement cc-step command handler with step bounds validation in commands/cc_step.go
- [ ] T023 [US2] Update playback loop to iterate step.CCValues and send all CC messages for the step in playback/playback.go
- [ ] T024 [US2] Implement cc-clear command parser in commands/cc_clear.go (parse: `cc-clear <step> [cc-number]`)
- [ ] T025 [US2] Implement cc-clear command handler (clear specific CC or all CCs from step) in commands/cc_clear.go
- [ ] T026 [US2] Implement cc-apply command parser in commands/cc_apply.go (parse: `cc-apply <cc-number>`)
- [ ] T027 [US2] Implement cc-apply command handler (copy globalCC value to all steps with notes) in commands/cc_apply.go
- [ ] T028 [US2] Add validation for cc-apply (error if no global value set for that CC number) in commands/cc_apply.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - users can create per-step automation and convert global to per-step

---

## Phase 5: User Story 3 - Save and Load CC Data with Patterns (Priority: P3)

**Goal**: Users can save patterns with CC automation and reload them with full CC data fidelity

**Independent Test**: Create pattern with notes and CC automation, save with `save filter-sweep`, load in new session with `load filter-sweep`, verify all CC automation plays correctly

### Implementation for User Story 3

- [ ] T029 [US3] Verify Step JSON marshal includes cc field (already done in T007) - validation task
- [ ] T030 [US3] Verify Step JSON unmarshal handles cc field (already done in T008) - validation task
- [ ] T031 [US3] Test backward compatibility by loading pattern created before CC feature in sequence/pattern_test.go
- [ ] T032 [US3] Implement save warning logic: check len(globalCC) > 0 in commands/save.go
- [ ] T033 [US3] Add warning message display when saving with unsaved global CC in commands/save.go
- [ ] T034 [US3] Update show command to display CC automation indicators for steps in commands/show.go
- [ ] T035 [US3] Test save/load round-trip with CC data (manual test with MIDI hardware)

**Checkpoint**: All user stories 1-3 should now be independently functional - patterns with CC automation persist correctly

---

## Phase 6: User Story 4 - Visual Feedback for CC Automation (Priority: P4)

**Goal**: Users can see which CC parameters are active on which steps to understand their automation

**Independent Test**: Create CC automation on multiple steps, run `cc-show`, verify table displays all active CC automations per step

### Implementation for User Story 4

- [ ] T036 [US4] Implement cc-show command parser in commands/cc_show.go
- [ ] T037 [US4] Implement cc-show display logic: iterate all steps and show CC values in table format in commands/cc_show.go
- [ ] T038 [US4] Handle empty CC automation case (display "No CC automation configured") in commands/cc_show.go
- [ ] T039 [US4] Enhance show command output to include CC indicators (e.g., [CC74:127]) in commands/show.go

**Checkpoint**: All 4 user stories complete - users have full CC control with visual feedback

---

## Phase 7: AI Integration (Cross-Cutting)

**Purpose**: Enable AI to generate patterns with CC automation

- [ ] T040 [P] Update AI system prompts to include cc command in ai/prompts.go or ai/system_prompt.go
- [ ] T041 [P] Update AI system prompts to include cc-step command in ai/prompts.go or ai/system_prompt.go
- [ ] T042 [P] Update AI system prompts to include cc-apply command in ai/prompts.go or ai/system_prompt.go
- [ ] T043 [P] Update AI system prompts to include cc-clear command in ai/prompts.go or ai/system_prompt.go
- [ ] T044 [P] Update AI system prompts to include cc-show command in ai/prompts.go or ai/system_prompt.go
- [ ] T045 Add CC command examples to AI prompt templates in ai/prompts.go or ai/system_prompt.go
- [ ] T046 Test AI mode: ask AI to "create a dark bass with filter sweep" and verify it generates CC commands

**Checkpoint**: AI can now generate patterns with CC automation

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and quality improvements

- [ ] T047 Manual test: Load pattern created before CC feature (backward compatibility)
- [ ] T048 Manual test: Global CC workflow (cc → play → hear change → cc-apply → save → load)
- [ ] T049 Manual test: Per-step automation workflow (cc-step on multiple steps → filter sweep)
- [ ] T050 Manual test: Save warning (set global CC → save → verify warning displayed)
- [ ] T051 Manual test: cc-show display (create automation → cc-show → verify table format)
- [ ] T052 Manual test: Edge cases (invalid CC number, step beyond length, multiple CC per step)
- [ ] T053 [P] Verify timing precision with MIDI monitor: CC messages within ±5ms of Note On
- [ ] T054 [P] Code review: Verify all CC state access is mutex-protected
- [ ] T055 [P] Code cleanup: Add code comments for CC data structures per data-model.md
- [ ] T056 Validate quickstart.md examples work as documented (run through all example sessions)
- [ ] T057 Update CLAUDE.md if any architectural changes were made during implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 2 uses User Story 1 concepts (global CC) but is independently testable
  - User Story 3 depends on User Story 2 (needs per-step CC to save)
  - User Story 4 is independent (just display/visual feedback)
- **AI Integration (Phase 7)**: Can proceed in parallel with user stories or after
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories ✅ MVP
- **User Story 2 (P2)**: Conceptually builds on US1 (global → per-step conversion) but independently testable
- **User Story 3 (P3)**: Depends on US2 (needs per-step CC automation to persist)
- **User Story 4 (P4)**: Independent - just display logic

### Within Each User Story

- Command parser before command handler
- Validation before state modification
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (different files)
- All AI Integration tasks marked [P] can run in parallel (different prompt sections)
- All Polish tasks marked [P] can run in parallel (different concerns)
- User Story 4 can be worked on in parallel with User Story 3 (independent)

---

## Parallel Example: User Story 1

```bash
# Within User Story 1, these tasks can run in parallel (different files):
Task T013: "Implement SetGlobalCC in sequence/sequence.go"
Task T014: "Implement GetGlobalCC in sequence/sequence.go"  # Same file as T013, must be sequential
Task T015: "Add cc command parser in commands/cc.go"       # Different file, can be parallel with T013/T014
```

Note: T013 and T014 are in the same file (sequence/sequence.go), so they must run sequentially. T015 is in a different file (commands/cc.go), so it can run in parallel with T013/T014.

---

## Parallel Example: AI Integration (Phase 7)

```bash
# All AI prompt updates can run in parallel (different prompt sections):
Task T040: "Update AI system prompts to include cc command"
Task T041: "Update AI system prompts to include cc-step command"
Task T042: "Update AI system prompts to include cc-apply command"
Task T043: "Update AI system prompts to include cc-clear command"
Task T044: "Update AI system prompts to include cc-show command"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (review existing codebase)
2. Complete Phase 2: Foundational (data structures - CRITICAL)
3. Complete Phase 3: User Story 1 (global CC commands)
4. **STOP and VALIDATE**: Test User Story 1 with MIDI hardware
5. Demo/iterate if ready

**Estimated MVP Scope**: T001-T018 (18 tasks) delivers working global CC control

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (T001-T012)
2. Add User Story 1 → Test independently → Demo (T013-T018) 🎯 **MVP!**
3. Add User Story 2 → Test independently → Demo (T019-T028)
4. Add User Story 3 → Test independently → Demo (T029-T035)
5. Add User Story 4 → Test independently → Demo (T036-T039)
6. Add AI Integration → Test with AI mode (T040-T046)
7. Polish and validate (T047-T057)

Each story adds creative value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T012)
2. Once Foundational is done:
   - Developer A: User Story 1 (T013-T018)
   - Developer B: User Story 4 (T036-T039) - independent
   - Developer C: AI Integration (T040-T046) - independent
3. After User Story 1 completes:
   - Developer A: User Story 2 (T019-T028)
4. After User Story 2 completes:
   - Developer A: User Story 3 (T029-T035)
5. Polish together (T047-T057)

---

## Task Count Summary

- **Phase 1 (Setup)**: 5 tasks
- **Phase 2 (Foundational)**: 7 tasks ← BLOCKING
- **Phase 3 (User Story 1 - P1)**: 6 tasks ← MVP
- **Phase 4 (User Story 2 - P2)**: 10 tasks
- **Phase 5 (User Story 3 - P3)**: 7 tasks
- **Phase 6 (User Story 4 - P4)**: 4 tasks
- **Phase 7 (AI Integration)**: 7 tasks
- **Phase 8 (Polish)**: 11 tasks

**Total**: 57 tasks

**MVP Scope** (P1 only): 18 tasks (Setup + Foundational + User Story 1)

**Parallel Opportunities**: 5 tasks in Setup, 12 tasks in AI/Polish, User Story 4 can overlap with User Story 3

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable with MIDI hardware
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently with synthesizer
- All file paths are relative to repository root
- No automated tests requested - manual MIDI hardware testing throughout
