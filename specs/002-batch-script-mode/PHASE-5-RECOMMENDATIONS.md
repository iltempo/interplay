# Phase 5 Recommendations: Lessons from Batch/Script Mode

**Based on**: Successful completion of Phase 4 (Batch/Script Mode)
**Date**: 2025-12-05
**For**: Phase 5 (MIDI CC Parameter Control) planning

## Key Learnings from Phase 4

### What Worked Extremely Well

1. **Clear User Stories with Priorities**:
   - P1/P2/P3 priority system helped focus on MVP first
   - Each story had independent test criteria
   - Sequential dependencies clearly documented

2. **Phased Task Breakdown**:
   - Setup → Foundation → US1 (MVP) → US2 → US3 → Polish
   - Clear checkpoints after each phase
   - Parallel task markers [P] enabled efficient execution

3. **Documentation-First Approach**:
   - README updates concurrent with implementation
   - Example files (6 test scripts) validated functionality
   - CLAUDE.md kept in sync with project evolution

4. **Runtime Validation Strategy**:
   - Simple approach: validate during execution via errors
   - Avoided complex pre-validation logic
   - Met all requirements with minimal code

5. **Enhancement Mindset**:
   - Script-to-interactive transition emerged during implementation
   - Recognized as valuable, documented, kept in scope
   - Didn't over-engineer; kept it simple

### What Could Be Improved

1. **Function Signature Evolution**:
   - Initial task specified `bool` return, evolved to `(bool, bool)`
   - Recommendation: Consider multi-value returns upfront in planning

2. **Implicit vs Explicit Requirements**:
   - "Result display" was implicit via command handlers
   - Could have been more explicit in spec to avoid ambiguity

3. **Validation Terminology**:
   - "Pre-execution validation" vs "runtime validation" caused confusion
   - Recommendation: Be more precise about validation strategy upfront

## Recommendations for Phase 5 (MIDI CC Control)

### Specification Phase

1. **Define CC Validation Strategy Early**:
   - How to validate CC# (0-127) and values (0-127)?
   - Runtime validation or pre-check?
   - What happens if invalid CC sent to synth?

2. **Clarify Data Model Immediately**:
   - How are CC values stored? (map[int]int per step?)
   - Global vs per-step CC (already partially implemented)
   - Persistence format in JSON

3. **Profile System Scope Boundaries**:
   - Phase 5a: Generic CC (✅ already clear)
   - Phase 5b: Profile loading (defer to future?)
   - Phase 5c: AI integration (defer to future?)
   - Phase 5d: Profile builder (separate project, defer)

4. **User Story Priorities**:
   - **P1 (MVP)**: Send global CC, persist to JSON
   - **P2**: Per-step CC automation
   - **P3**: Multiple CC per step
   - **P4**: Visual feedback (`cc-show`)

### Planning Phase

1. **Break Down by User Story** (like Phase 4):
   ```
   Phase 1: Setup (dependencies)
   Phase 2: Foundation (CC data model, JSON persistence)
   Phase 3: US1 - Global CC (MVP)
   Phase 4: US2 - Per-Step CC
   Phase 5: US3 - Multiple CC per Step
   Phase 6: US4 - Visual Feedback
   Phase 7: Polish (docs, examples, validation)
   ```

2. **Example Scripts Early**:
   - Create `test_cc_global.txt`, `test_cc_sweep.txt`, etc.
   - Use these to validate functionality throughout
   - Include in spec as acceptance criteria

3. **Document JSON Format Upfront**:
   ```json
   {
     "step": 1,
     "note": "C3",
     "cc": {
       "74": 127,  // Filter cutoff
       "71": 64    // Resonance
     }
   }
   ```

4. **AI Integration Strategy**:
   - How does AI know which CC to use?
   - "Make it darker" → CC#74 (filter) = lower value
   - Document heuristics before implementation

### Implementation Phase

1. **Start with Tests**:
   - Create `test_cc_basic.txt` first
   - Validate each command works before moving on
   - Manual testing is fine (like Phase 4)

2. **Incremental Commits**:
   - Each user story = separate commit
   - Easy to review, easy to revert if needed
   - Document enhancements as they emerge

3. **Keep TODO List Active**:
   - Use TodoWrite tool to track phase progress
   - Mark tasks complete immediately
   - Helps maintain momentum

4. **Profile System Deferral**:
   - Generic CC (Phase 5a) is sufficient for MVP
   - Profile system (5b/5c/5d) can be separate feature
   - Don't over-engineer; deliver value fast

### Documentation Phase

1. **Update CLAUDE.md Immediately**:
   - Mark Phase 5a complete when done
   - Document CC command syntax
   - Add usage examples

2. **README Section Structure**:
   ```markdown
   ## MIDI CC Control

   ### Basic Usage
   ### Per-Step Automation
   ### Multiple Parameters
   ### Synth-Specific Tips
   ### Example Scripts
   ```

3. **Create Implementation Summary**:
   - Like `IMPLEMENTATION-SUMMARY.md` for Phase 4
   - Document learnings, commits, validation
   - Reference for future phases

## Specific Phase 5 Considerations

### Data Model Extension

Current `Step` structure:
```go
type Step struct {
    Note       *int
    Velocity   int
    Gate       int
    CCValues   map[int]int  // Already exists!
}
```

**Recommendation**: CCValues already implemented! Verify current state before planning new work.

### Command Design

Proposed commands (verify against existing):
- `cc <number> <value>` - Global CC (transient)
- `cc-step <step> <number> <value>` - Per-step CC (persistent)
- `cc-clear <step> <number>` - Remove CC from step
- `cc-show` - Display all active CC

**Check**: Are these already implemented? Review `commands/cc*.go` files before planning.

### JSON Persistence

Verify current pattern save/load includes CCValues map. If not, this is the core work for Phase 5a.

### AI Integration

Current AI handler architecture:
- AI generates command strings
- Commands execute via existing handlers
- CC commands should work automatically

**Test**: Can AI already say "set CC 74 to 64" and have it work?

## Action Items Before Starting Phase 5

1. ✅ **Audit Existing CC Implementation**:
   - Review `commands/cc*.go` files
   - Check what's already done
   - Identify gaps vs Phase 5a requirements

2. ✅ **Test Current CC Functionality**:
   - Try `cc 74 127` command
   - Try `cc-step 1 74 127` command
   - Check if JSON persistence works

3. ✅ **Define Phase 5a Scope Precisely**:
   - List only missing features
   - Don't duplicate existing work
   - Focus on gaps, not reimplementation

4. ✅ **Create Phase 5a Spec**:
   - Use Phase 4 as template
   - User stories with priorities
   - Clear acceptance criteria
   - Manual test examples

5. ✅ **Run `/speckit.specify` for Phase 5a**:
   - Generate spec.md from feature description
   - Clarify ambiguities before planning
   - Run `/speckit.analyze` before implementation

## Timeline Estimate

Based on Phase 4 (30 tasks, ~4 hours):

- **Phase 5a (Generic CC)**: ~15 tasks, 2-3 hours (if starting from scratch)
- **If CC partially implemented**: ~5-10 tasks, 1-2 hours
- **Phase 5b (Profiles)**: Defer to separate feature (30+ tasks)
- **Phase 5c (AI integration)**: Defer or test if already working

## Success Criteria Template

Copy from Phase 4, adapt for CC:

- **SC-001**: Users can set global CC values that persist until changed
- **SC-002**: Users can automate CC per step for filter sweeps, etc.
- **SC-003**: Multiple CC parameters can control same step
- **SC-004**: CC values persist in saved patterns (JSON)
- **SC-005**: `cc-show` displays all active CC automations
- **SC-006**: AI commands can manipulate CC values

## Final Recommendations

1. **Audit First**: Check what's already implemented before planning
2. **MVP Focus**: Generic CC (Phase 5a) is sufficient for initial release
3. **Defer Profiles**: Phase 5b/5c/5d are separate features, not blockers
4. **Test-Driven**: Create example scripts before implementation
5. **Document As You Go**: Keep CLAUDE.md and README in sync
6. **Commit Incrementally**: Each user story = one commit
7. **Celebrate Wins**: Mark tasks complete immediately
8. **Learn & Adapt**: Phase 4 patterns work well, reuse them

---

**Next Step**: Run audit of existing CC implementation, then create Phase 5a spec with `/speckit.specify`.
