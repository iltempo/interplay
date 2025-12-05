# Data Model: MIDI CC Parameter Control

**Feature**: 001-midi-cc-control
**Date**: 2025-12-04
**Status**: Complete

## Overview

This document defines the data structures for MIDI CC (Control Change) parameter control in Interplay. The data model extends existing pattern structures to support both global (transient) and per-step (persistent) CC automation.

## Core Entities

### Entity 1: Step (Extended)

**Purpose**: Represents a single step in a pattern, now including optional CC automation

**Attributes**:
- `Note`: Optional MIDI note number (nil = rest) - **EXISTING**
- `Velocity`: Note velocity 0-127 - **EXISTING**
- `Gate`: Gate length percentage 1-100 - **EXISTING**
- `CCValues`: Map of CC automation (CC number → value) - **NEW**

**Go Structure**:
```go
type Step struct {
    Note       *int           // nil for rest, MIDI note 0-127
    Velocity   int            // 0-127
    Gate       int            // 1-100 percentage
    CCValues   map[int]int    // CC# → Value (0-127)
}
```

**JSON Representation**:
```json
{
  "step": 1,
  "note": "C3",
  "velocity": 100,
  "gate": 90,
  "cc": {
    "74": 127,
    "71": 64
  }
}
```

**Validation Rules**:
- CC number must be 0-127 (standard MIDI range)
- CC value must be 0-127 (standard MIDI range)
- CCValues map may be nil or empty (no CC automation)
- CCValues not persisted if empty (omitempty in JSON)

**State Transitions**:
- Initial: CCValues = nil (no automation)
- Add CC: CCValues[ccNum] = value (creates map if nil)
- Remove CC: delete(CCValues, ccNum) (map remains, may be empty)
- Clear all CC: CCValues = nil or CCValues = make(map[int]int)

### Entity 2: Sequence (Extended)

**Purpose**: Represents the entire pattern state, including transient global CC values

**Attributes** (showing CC-related only):
- `steps`: Slice of Step structs - **EXISTING**
- `globalCC`: Map of transient global CC values - **NEW**

**Go Structure**:
```go
type Sequence struct {
    mu           sync.Mutex
    steps        []Step
    length       int
    tempo        int
    // ... other existing fields

    globalCC     map[int]int    // NEW: transient global CC state
}
```

**Validation Rules**:
- globalCC not persisted to JSON (transient state)
- globalCC CC numbers must be 0-127
- globalCC values must be 0-127
- globalCC warns user on save if non-empty

**State Transitions**:
- Initial: globalCC = nil
- Set global: globalCC[ccNum] = value
- Apply to steps: copy globalCC[ccNum] to all steps with notes
- Load pattern: globalCC = nil (not restored)

### Entity 3: Pattern JSON (Extended)

**Purpose**: Persistent pattern representation with CC automation data

**Structure**:
```json
{
  "name": "Dark Filter Sweep",
  "tempo": 80,
  "length": 16,
  "humanization": {
    "velocity": 8,
    "timing": 10,
    "gate": 5
  },
  "swing": 0,
  "steps": [
    {
      "step": 1,
      "note": "C3",
      "velocity": 100,
      "gate": 90,
      "cc": {
        "74": 127,
        "71": 64
      }
    },
    {
      "step": 5,
      "note": "G2",
      "velocity": 80,
      "gate": 90,
      "cc": {
        "74": 20
      }
    }
  ]
}
```

**Validation Rules**:
- All existing pattern validation rules apply
- `cc` field is optional per step (omitted if no CC automation)
- CC keys are strings in JSON (converted to int in Go)
- Backward compatible: patterns without `cc` field load successfully

## Relationships

### Step ↔ CC Automation

**Relationship**: One-to-Many (one step can have multiple CC parameters)

**Cardinality**: 0..N (step may have zero or unlimited CC automations)

**Implementation**: Map structure allows sparse storage and efficient lookup

**Example**:
```
Step 1:
  Note: C3
  CCValues:
    74 → 127  (filter cutoff fully open)
    71 → 64   (resonance mid-level)
    73 → 10   (envelope attack fast)
```

### Sequence ↔ Steps

**Relationship**: One-to-Many (existing relationship, unchanged)

**CC Impact**: Each step now optionally contains CC automation

**Playback**: Iterate steps, send CC messages before Note On

### Global CC ↔ Per-Step CC

**Relationship**: Conversion (global CC can be applied to steps)

**Semantics**:
- Global CC is transient (live experimentation)
- Per-step CC is persistent (saved automation)
- `cc-apply` converts global → per-step for all steps with notes

**Example Conversion**:
```
Before cc-apply 74:
  globalCC: {74: 100}
  Step 1: note=C3, cc={}
  Step 5: note=G2, cc={}

After cc-apply 74:
  globalCC: {74: 100}  (unchanged, still live)
  Step 1: note=C3, cc={74: 100}
  Step 5: note=G2, cc={74: 100}
```

## Data Flow

### User Command → Pattern State → MIDI Out

```
1. User: cc-step 1 74 127

2. Command Parser:
   Parse: step=1, cc=74, value=127
   Validate: 1 ≤ step ≤ length, 0 ≤ cc ≤ 127, 0 ≤ value ≤ 127

3. Sequence State (mutex protected):
   sequence.steps[0].CCValues[74] = 127

4. Playback (next loop iteration):
   For step 1:
     - Send CC#74 value 127 (MIDI: 0xB0 0x4A 0x7F)
     - Send Note On C3

5. Pattern Display:
   Step 1: C3 v:100 g:90 [CC74:127]
```

### Pattern Save → JSON File

```
1. User: save my-pattern

2. Check Global CC:
   if len(globalCC) > 0:
     warn: "Global CC values will not be saved. Use cc-apply."

3. Marshal to JSON:
   For each step:
     if len(step.CCValues) > 0:
       include "cc" field
     else:
       omit "cc" field (omitempty)

4. Write File:
   patterns/my-pattern.json
```

### Pattern Load → Pattern State

```
1. User: load my-pattern

2. Read JSON File:
   Unmarshal into Pattern struct

3. Restore State:
   For each step in JSON:
     step.Note = parse note string
     step.Velocity = velocity (default 100)
     step.Gate = gate (default 90)
     step.CCValues = cc map (or nil if absent)

4. Initialize Transient State:
   sequence.globalCC = nil (not persisted)
```

## Constraints and Invariants

### Constraint 1: MIDI CC Range

**Rule**: All CC numbers and values MUST be 0-127

**Enforcement**: Validation in command parser before state update

**Error Handling**: Display error message, do not modify state

### Constraint 2: Global CC Not Persisted

**Rule**: globalCC map MUST NOT be included in JSON serialization

**Enforcement**: Not included in JSON struct tags, warnings on save

**Rationale**: Global CC is for live experimentation, not permanent automation

### Constraint 3: Thread Safety

**Rule**: All CC state access MUST be mutex-protected

**Enforcement**: Use sequence.mu lock for all reads/writes

**Rationale**: Pattern state shared between command handler and playback goroutine

### Constraint 4: Timing Precision

**Rule**: CC messages MUST be sent within ±5ms of note messages

**Enforcement**: Send all CC messages synchronously before Note On

**Rationale**: Parameter must be set before note triggers for correct sound

### Constraint 5: Backward Compatibility

**Rule**: Patterns without CC data MUST load without errors

**Enforcement**: Use omitempty JSON tags, nil-safe map operations

**Verification**: Test suite includes patterns from before CC feature

## Migration Strategy

### No Migration Required

**Rationale**:
- New field is optional (`omitempty` tag)
- Go handles missing JSON fields gracefully (nil map)
- No schema version needed

**Verification Steps**:
1. Load pattern created before CC feature
2. Verify no errors or warnings
3. Verify pattern plays correctly (no CC messages)
4. Add CC automation to loaded pattern
5. Save and reload
6. Verify CC automation persists

## Example Scenarios

### Scenario 1: Simple Filter Sweep

**User Actions**:
```
> set 1 C2
> set 5 G2
> set 9 C2
> set 13 F2
> cc-step 1 74 127    # Filter open
> cc-step 5 74 80
> cc-step 9 74 40
> cc-step 13 74 60
> save filter-sweep
```

**Data Model State**:
```go
sequence.steps[0] = Step{
    Note: ptr(36),  // C2
    Velocity: 100,
    Gate: 90,
    CCValues: map[int]int{74: 127},
}
// ... similar for steps 4, 8, 12
```

**JSON Output**:
```json
{
  "name": "filter-sweep",
  "steps": [
    {"step": 1, "note": "C2", "cc": {"74": 127}},
    {"step": 5, "note": "G2", "cc": {"74": 80}},
    {"step": 9, "note": "C2", "cc": {"74": 40}},
    {"step": 13, "note": "F2", "cc": {"74": 60}}
  ]
}
```

### Scenario 2: Global CC Experimentation

**User Actions**:
```
> load my-pattern
> cc 74 50         # Experiment with filter
> cc 71 80         # Experiment with resonance
> save my-pattern
Warning: Global CC values (CC#74, CC#71) will not be saved. Use 'cc-apply 74' to make permanent.
```

**Data Model State**:
```go
sequence.globalCC = map[int]int{
    74: 50,
    71: 80,
}
// Steps unchanged - global CC not applied
```

### Scenario 3: Convert Experiment to Permanent

**User Actions**:
```
> cc 74 100        # Find a good filter setting
> cc-apply 74      # Apply to all steps with notes
> save my-pattern
Pattern saved: my-pattern.json
```

**Data Model Transformation**:
```go
// Before cc-apply:
globalCC: {74: 100}
steps[0]: {note: C2, cc: {}}
steps[4]: {note: G2, cc: {}}

// After cc-apply:
globalCC: {74: 100}  // Still present for live tweaking
steps[0]: {note: C2, cc: {74: 100}}  // Now persistent
steps[4]: {note: G2, cc: {74: 100}}  // Now persistent
```

## Implementation Checklist

- [ ] Add CCValues map[int]int to Step struct
- [ ] Add globalCC map[int]int to Sequence struct
- [ ] Update JSON marshal/unmarshal with omitempty
- [ ] Add CC validation (0-127 range)
- [ ] Add mutex protection for CC state access
- [ ] Implement global CC warning on save
- [ ] Test backward compatibility with old patterns
- [ ] Verify timing precision (<±5ms)
- [ ] Document CC data model in code comments
