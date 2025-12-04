# Research: MIDI CC Parameter Control

**Feature**: 001-midi-cc-control
**Date**: 2025-12-04
**Status**: Complete

## Overview

This document consolidates research findings for implementing generic MIDI CC (Control Change) parameter control in Interplay. Since the technical context is well-defined and no NEEDS CLARIFICATION items exist, this research focuses on best practices, implementation patterns, and design decisions.

## Key Design Decisions

### Decision 1: Global vs Per-Step CC Persistence

**Decision**: Global CC commands (`cc`) are transient (not saved), per-step CC automation (`cc-step`) is persisted

**Rationale**:
- Clear mental model: `cc` = live experimentation, `cc-step` = permanent automation
- Aligns with pattern-based workflow (patterns contain automation, not live state)
- Prevents user confusion about what's temporary vs permanent
- Provides explicit conversion path via `cc-apply` command

**Alternatives Considered**:
- Save everything (global + per-step): Rejected - confusing semantics, unclear what's automation vs live state
- Make it user-configurable with `--persist` flag: Rejected - adds complexity, violates simplicity principle
- Never save anything: Rejected - defeats purpose of pattern persistence

### Decision 2: CC Data Structure in Step

**Decision**: Use `map[int]int` to store CC automation per step (CC number → value)

**Rationale**:
- Go idiomatic: maps are efficient for sparse key-value data
- Supports unlimited CC parameters per step without fixed arrays
- Easy JSON serialization/deserialization
- Simple lookup and modification operations

**Implementation**:
```go
type Step struct {
    Note       *int           // Existing
    Velocity   int            // Existing
    Gate       int            // Existing
    CCValues   map[int]int    // NEW: CC# → Value
}
```

**Alternatives Considered**:
- Array `[128]int`: Rejected - wastes memory, most steps use 0-3 CC parameters
- Slice of struct `[]CCAutomation`: Rejected - more complex, harder to look up/modify
- Separate CC automation data structure: Rejected - complicates data model, breaks cohesion

### Decision 3: MIDI CC Message Timing

**Decision**: Send CC messages at step boundaries alongside Note On messages

**Rationale**:
- Consistent with existing note/velocity/gate timing model
- No additional timing complexity or goroutines needed
- Maintains ±5ms timing precision guarantee
- CC messages sent before Note On to ensure parameter is set when note triggers

**Message Order**:
1. CC messages for step (all CC values in map)
2. Note On (if step has note)
3. Schedule Note Off based on gate

**Alternatives Considered**:
- Send CC messages mid-step (interpolation): Rejected - violates pattern-based simplicity, adds timing complexity
- Separate CC timing goroutine: Rejected - unnecessary complexity, harder to maintain timing guarantees
- CC messages after Note On: Rejected - parameter might not be set when note triggers

### Decision 4: Backward Compatibility Strategy

**Decision**: Pattern JSON includes optional `cc` field, omitted if no CC automation

**Rationale**:
- Old patterns (no `cc` field) load without errors - field simply absent
- New patterns with CC save `cc` object per step
- JSON remains human-readable and manually editable
- No migration scripts needed

**JSON Format**:
```json
{
  "step": 1,
  "note": "C3",
  "velocity": 100,
  "gate": 90,
  "cc": {              // NEW - optional field
    "74": 127,         // Filter cutoff
    "71": 64           // Resonance
  }
}
```

**Alternatives Considered**:
- Version field in JSON: Rejected - unnecessary, Go's JSON unmarshal handles missing fields
- Separate CC file: Rejected - splits pattern data, complicates persistence
- Always include empty `cc: {}`: Rejected - clutters JSON for patterns without CC

### Decision 5: Global CC State Management

**Decision**: Track global CC values in separate `globalCC map[int]int` in sequence state

**Rationale**:
- Separation of concerns: global (transient) vs per-step (persistent)
- Easy to detect unsaved global CC for save warning
- Simple to implement `cc-apply` conversion
- No impact on existing pattern state

**Warning Implementation**:
```go
// On save command
if len(sequence.GlobalCC) > 0 {
    fmt.Println("Warning: Global CC values (...) will not be saved.")
    fmt.Println("Use 'cc-apply <number>' to make permanent.")
}
```

**Alternatives Considered**:
- Store global CC as "step 0": Rejected - confusing, breaks step numbering semantics
- No global CC state (always per-step): Rejected - requires setting CC on every step manually
- Auto-convert on save: Rejected - surprising behavior, violates user control principle

## Go Implementation Patterns

### Pattern 1: Thread-Safe CC State Access

**Best Practice**: Use existing mutex to protect CC state access

```go
// In sequence package
type Sequence struct {
    mu         sync.Mutex
    steps      []Step
    globalCC   map[int]int    // NEW: transient global CC state
    // ... existing fields
}

func (s *Sequence) SetGlobalCC(ccNumber, value int) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.globalCC == nil {
        s.globalCC = make(map[int]int)
    }
    s.globalCC[ccNumber] = value
}
```

**Rationale**: Extends existing pattern state mutex protection, no new synchronization primitives needed

### Pattern 2: MIDI CC Message Sending

**Best Practice**: Use existing `midi` package Send method with CC message type

```go
// In midi package
func (m *MIDIOut) SendCC(channel, ccNumber, value uint8) error {
    return midi.Send(m.out, midi.ControlChange(channel, ccNumber, value))
}
```

**Rationale**: Leverages existing MIDI library, consistent with note sending pattern

### Pattern 3: JSON Marshaling with Omitempty

**Best Practice**: Use `omitempty` tag for optional CC field

```go
type StepJSON struct {
    Step     int            `json:"step"`
    Note     *string        `json:"note,omitempty"`
    Velocity int            `json:"velocity,omitempty"`
    Gate     int            `json:"gate,omitempty"`
    CC       map[string]int `json:"cc,omitempty"`  // NEW: string keys for JSON
}
```

**Rationale**: Clean JSON output, backward compatible, follows Go conventions

## Testing Strategy

### Unit Testing

**Scope**:
- CC state management (set, get, clear operations)
- CC-apply conversion logic
- JSON marshaling/unmarshaling with CC data
- Backward compatibility (load old patterns)

**Approach**: Standard Go `testing` package, table-driven tests

### Integration Testing

**Scope**:
- End-to-end command flow (cc → state → MIDI out)
- Pattern save/load with CC data
- Playback with CC automation

**Approach**: Test helpers with mock MIDI output

### Manual Testing

**Scope**:
- Actual MIDI hardware verification
- Timing precision validation (oscilloscope/MIDI monitor)
- User workflow testing (experiment → convert → save)

**Required Equipment**: MIDI synthesizer (Waldorf Robot or similar), MIDI monitoring software

## Performance Considerations

### Memory Impact

**Current**: ~100 bytes per pattern (16 steps × ~6 bytes per step)
**With CC**: ~200-400 bytes per pattern (assuming 2-4 CC params per step)

**Impact**: Negligible - patterns are small, memory not a constraint

### CPU Impact

**Additional Operations**:
- Map lookups per step (O(1), ~10ns)
- Extra MIDI messages per step (2-4 CC messages = ~1-2µs)

**Impact**: Negligible - well within ±5ms timing budget

### I/O Impact

**MIDI Bandwidth**:
- Note On/Off: 6 bytes per step
- CC messages: 3 bytes each × 2-4 = 6-12 bytes per step
- Total: 12-18 bytes per step (vs 6 bytes current)

**MIDI Bandwidth**: 3125 bytes/sec at 31.25 kbaud
**Pattern Bandwidth**: ~288 bytes/sec at 80 BPM, 16 steps
**Utilization**: <10% of MIDI bandwidth

**Impact**: No bandwidth concerns, no timing impact

## Dependencies and Integration Points

### Module: `sequence/`

**Changes**:
- Add `CCValues map[int]int` to Step struct
- Add `globalCC map[int]int` to Sequence state
- Add CC getter/setter methods
- Update JSON marshal/unmarshal

**Impact**: Isolated to sequence package, no breaking changes to existing API

### Module: `midi/`

**Changes**:
- Add `SendCC(channel, ccNumber, value uint8)` method

**Impact**: Single new method, no changes to existing MIDI functionality

### Module: `playback/`

**Changes**:
- Send CC messages before Note On at each step
- Iterate over step.CCValues map and call SendCC for each entry

**Impact**: Small addition to existing step playback loop (~5 lines)

### Module: `commands/`

**Changes**:
- Add parsers for: `cc`, `cc-step`, `cc-apply`, `cc-clear`, `cc-show`
- Add save warning logic for unsaved global CC

**Impact**: New command handlers, no changes to existing commands

### Module: `ai/`

**Changes**:
- Update system prompt strings to include CC commands
- Add CC command examples to prompt templates

**Impact**: String updates only, no code logic changes

## Risk Assessment

### Risk 1: MIDI Timing Precision

**Risk**: CC messages might affect note timing if not implemented carefully

**Mitigation**:
- Send all CC messages for a step before Note On
- Pre-calculate all MIDI messages before step boundary
- Test with MIDI monitor and hardware to verify <±5ms tolerance

**Severity**: Medium | **Likelihood**: Low | **Status**: Mitigated

### Risk 2: Backward Compatibility

**Risk**: Old patterns might fail to load or behave incorrectly

**Mitigation**:
- Use `omitempty` JSON tags
- Test loading patterns created before CC feature
- Validate no errors or warnings on old pattern load

**Severity**: High | **Likelihood**: Very Low | **Status**: Mitigated

### Risk 3: User Confusion (Global vs Per-Step)

**Risk**: Users might not understand transient vs persistent CC semantics

**Mitigation**:
- Clear warning message when saving with unsaved global CC
- Quickstart documentation explains workflow
- `cc-apply` command makes conversion explicit and easy

**Severity**: Medium | **Likelihood**: Medium | **Status**: Mitigated

## Open Questions

**None** - all design decisions finalized through clarification session and research.

## Next Steps

Proceed to **Phase 1: Design & Contracts**:
1. Create detailed data model (data-model.md)
2. Generate quickstart guide (quickstart.md)
3. Update agent context with Go/MIDI patterns
4. Validate design against constitution (re-check gates)
