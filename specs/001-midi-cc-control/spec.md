# Feature Specification: MIDI CC Parameter Control

**Feature Branch**: `001-midi-cc-control`
**Created**: 2025-12-04
**Status**: Draft
**Input**: User description: "Add generic MIDI CC (Control Change) parameter control to Interplay. Users should be able to send any CC message (0-127) with any value (0-127), automate CC values per step like velocity and gate, save CC data with patterns, and control synthesizer parameters without needing synth-specific profiles. This provides the foundation for future AI-powered sound design."

## Clarifications

### Session 2025-12-04

- Q: When a user sends a global CC command (e.g., `cc 74 100`), should this value persist when the pattern is saved and loaded? → A: Global CC commands are transient (not saved with patterns). Only per-step CC automation (`cc-step`) is persisted. Provide easy conversion command (`cc-apply`) to make global CC permanent. Warn users when saving if global CC values exist that won't be saved.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send Global CC Messages (Priority: P1)

A user wants to quickly adjust a synthesizer parameter (like filter cutoff or resonance) that affects the entire pattern without automating it per step.

**Why this priority**: This is the simplest form of CC control and provides immediate value. Users can experiment with synth parameters in real-time while a pattern loops. This MVP delivers core functionality and validates the MIDI CC implementation.

**Independent Test**: Can be fully tested by connecting any MIDI synth, sending a CC command (e.g., `cc 74 100`), and verifying the synth responds by changing the parameter (e.g., filter opens). Delivers immediate creative control over synth sound.

**Acceptance Scenarios**:

1. **Given** a running pattern with notes, **When** user sends `cc 74 127`, **Then** filter cutoff (CC#74) is set to maximum on the connected synthesizer
2. **Given** a pattern playing, **When** user sends `cc 71 64`, **Then** resonance (CC#71) is set to mid-level on the synthesizer
3. **Given** multiple CC commands sent, **When** pattern loops, **Then** all CC values persist across loop iterations (but are not saved to pattern file)
4. **Given** an invalid CC number (> 127), **When** user sends command, **Then** system displays error message and does not send MIDI message
5. **Given** global CC value set with `cc 74 100`, **When** user saves and loads pattern, **Then** CC value resets to default (global CC is transient)

---

### User Story 2 - Per-Step CC Automation (Priority: P2)

A user wants to create dynamic parameter changes throughout the pattern, like a filter sweep where different steps have different filter cutoff values.

**Why this priority**: This unlocks creative sound design possibilities. Users can create movement and variation in their patterns beyond just note changes. This builds on P1's foundation and adds significant creative value.

**Independent Test**: Can be tested by setting different CC values for different steps (e.g., step 1: filter open, step 5: filter closed), playing the pattern, and hearing the filter sweep. Delivers automated parameter modulation.

**Acceptance Scenarios**:

1. **Given** a pattern with notes on steps 1 and 5, **When** user sets `cc-step 1 74 127` and `cc-step 5 74 20`, **Then** filter sweeps from open to closed between steps
2. **Given** a step with a note, **When** user assigns multiple CC values to the same step (filter + resonance), **Then** both CC messages are sent when that step plays
3. **Given** CC automation on a step, **When** user removes the note from that step, **Then** CC messages still send (allows parameter-only steps without notes)
4. **Given** CC automation on step 3, **When** user sends `cc-clear 3 74`, **Then** CC#74 automation is removed from step 3 only
5. **Given** global CC set with `cc 74 100`, **When** user sends `cc-apply 74`, **Then** CC#74 value (100) is applied to all steps with notes as per-step automation

---

### User Story 3 - Save and Load CC Data with Patterns (Priority: P3)

A user creates a pattern with both notes and CC automation (e.g., a bass line with a filter sweep), saves it, and wants to recall it later with all CC data intact.

**Why this priority**: Pattern persistence is essential for creative workflow. Users should be able to save their sonic creations and recall them in future sessions. This completes the feature by making CC data persistent.

**Independent Test**: Can be tested by creating a pattern with notes and CC automation, saving it with `save filter-sweep`, loading it in a new session with `load filter-sweep`, and verifying all CC automation plays correctly.

**Acceptance Scenarios**:

1. **Given** a pattern with notes and CC automation, **When** user saves with `save my-pattern`, **Then** both notes and CC data are saved to JSON file
2. **Given** a saved pattern with CC data, **When** user loads with `load my-pattern`, **Then** all CC automation is restored and plays correctly
3. **Given** a pattern created before CC feature existed (no CC data), **When** user loads it, **Then** pattern plays normally without CC messages (backward compatibility)
4. **Given** CC automation in a pattern, **When** user runs `show` command, **Then** display includes CC automation information for each step
5. **Given** global CC values set (e.g., `cc 74 100`) but not converted to per-step automation, **When** user saves pattern, **Then** system displays warning: "Warning: Global CC values (CC#74) will not be saved. Use 'cc-apply 74' to make permanent."

---

### User Story 4 - Visual Feedback for CC Automation (Priority: P4)

A user wants to see which CC parameters are active on which steps to understand and edit their parameter automation.

**Why this priority**: While not essential for functionality, visual feedback greatly improves usability. Users can see their automation at a glance rather than trying to remember which steps have which CC values.

**Independent Test**: Can be tested by creating CC automation and running `cc-show` command to display a table of all active CC automations per step.

**Acceptance Scenarios**:

1. **Given** CC automation on multiple steps, **When** user runs `cc-show`, **Then** display shows a table with step numbers, CC numbers, and values
2. **Given** no CC automation in pattern, **When** user runs `cc-show`, **Then** display shows "No CC automation configured"
3. **Given** CC automation exists, **When** user runs `show` command, **Then** pattern display includes indicator for steps with CC automation

---

### Edge Cases

- What happens when a user sends a CC message during pattern playback? (CC should queue and apply at next loop boundary, consistent with note changes)
- What happens when a user assigns CC automation to a step beyond the current pattern length? (System should display error: "Step X is beyond pattern length Y")
- What happens when multiple CC messages for the same CC number are assigned to one step? (Last value wins, similar to how only one note per step is allowed)
- What happens when loading a malformed pattern file with invalid CC data? (System should skip invalid CC entries, log warning, load valid portions)
- What happens when the MIDI connection drops during playback with CC automation? (Same error handling as notes - display error, attempt reconnection)
- What happens when `cc-apply` is used on a CC number that hasn't been set globally? (System should display error: "No global value set for CC#X")
- What happens when `cc-apply` is used and some steps already have automation for that CC number? (Overwrite existing values with the global value)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept CC commands with CC number (0-127) and value (0-127)
- **FR-002**: System MUST send MIDI CC messages to the connected synthesizer in real-time
- **FR-003**: System MUST allow per-step CC automation where each step can have multiple CC parameters
- **FR-004**: System MUST persist CC automation data when saving patterns to JSON files
- **FR-005**: System MUST restore CC automation data when loading patterns from JSON files
- **FR-006**: System MUST maintain backward compatibility with patterns created before CC feature (load without errors)
- **FR-007**: System MUST apply CC changes at loop boundaries, consistent with pattern-based synchronization model
- **FR-008**: System MUST send CC messages at the same timing precision as note messages (16th-note resolution)
- **FR-009**: System MUST validate CC numbers and values are within valid MIDI range (0-127)
- **FR-010**: System MUST display error messages for invalid CC commands without disrupting playback
- **FR-011**: System MUST provide visual feedback showing active CC automation per step
- **FR-012**: System MUST allow removal of CC automation from specific steps
- **FR-013**: System MUST support CC-only steps (no note, just parameter changes)
- **FR-014**: Users MUST be able to see current CC automation with `cc-show` command
- **FR-015**: Users MUST be able to see CC automation indicators in pattern display (`show` command)
- **FR-016**: System MUST NOT persist global CC values when saving patterns (only per-step CC automation is saved)
- **FR-017**: System MUST provide a conversion command (`cc-apply <cc-number>`) to convert current global CC value to per-step automation for all steps with notes
- **FR-018**: System MUST warn users when saving a pattern if global CC values exist that will not be persisted, suggesting `cc-apply` to make them permanent
- **FR-019**: AI system prompts MUST be updated to include all CC commands (`cc`, `cc-step`, `cc-apply`, `cc-clear`, `cc-show`) so AI can generate patterns with CC automation

### Key Entities

- **CC Automation**: Represents a Control Change message assigned to a specific step. Contains CC number (0-127), value (0-127), and step number (1-16+). Multiple CC automations can exist per step.

- **Pattern**: Extended to include CC automation data in addition to existing note, velocity, and gate information. Each step can now have an optional collection of CC automations.

- **MIDI Message**: Extended to include CC messages in addition to Note On/Off messages. CC messages are sent at step boundaries alongside note messages with same timing guarantees.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can send a global CC command and hear the synthesizer parameter change within one loop iteration (< 2 seconds at 80 BPM)
- **SC-002**: Users can create per-step CC automation (e.g., 8-step filter sweep) and hear smooth parameter modulation during playback
- **SC-003**: Users can save and load patterns with CC automation, with 100% data fidelity (all CC values restored correctly)
- **SC-004**: System maintains timing precision: CC messages are sent within ±5ms of note messages at step boundaries
- **SC-005**: Users can configure CC automation on a 16-step pattern in under 1 minute using command interface
- **SC-006**: System handles edge cases gracefully: invalid CC numbers/values show error messages without disrupting playback
- **SC-007**: Visual feedback displays all active CC automation clearly, allowing users to understand their automation at a glance
- **SC-008**: AI can generate patterns with CC automation when asked (e.g., "create a dark bass with filter sweep" produces both notes and CC commands)

## Assumptions

- Users understand basic MIDI CC concepts (CC numbers correspond to synth parameters)
- Users have access to their synthesizer's MIDI implementation chart to know which CC numbers control which parameters
- Future synth profile system will provide friendly parameter names (e.g., "filter" instead of "CC#74"), but this feature works without profiles
- CC messages follow standard MIDI specification (controller number 0-127, value 0-127)
- Pattern-based loop boundary synchronization applies to CC changes (no immediate real-time CC updates mid-pattern)
- JSON pattern files remain human-readable and manually editable if needed
- Global CC commands (`cc`) are for live experimentation (transient), per-step CC automation (`cc-step`) is for permanent sound design (persisted)
- Users will use `cc-apply` to convert successful experiments into permanent automation

## Dependencies

- Existing MIDI connection and message sending infrastructure (`midi/` module)
- Existing pattern state management and loop boundary synchronization (`sequence/`, `playback/` modules)
- Existing pattern persistence system (JSON save/load)
- Existing command parser framework (`commands/` module)
- Existing AI integration system (`ai/` module) - requires system prompt updates to include CC commands

## Out of Scope

- Synth-specific parameter naming (e.g., "filter" instead of "CC#74") - deferred to Phase 4b
- AI-powered sound design suggestions - deferred to Phase 4c
- MIDI learn functionality (auto-detect CC numbers from hardware) - future enhancement
- CC curve automation (linear interpolation between values) - future enhancement
- NRPN (Non-Registered Parameter Numbers) support - future enhancement
- SysEx (System Exclusive) messages - out of scope for this feature
