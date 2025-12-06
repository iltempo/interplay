# Feature Specification: Model Comparison Framework

**Feature Branch**: `003-model-comparison`
**Created**: 2025-12-06
**Status**: Draft
**Input**: User description: "Model comparison framework for systematically comparing AI model outputs for pattern generation quality"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Comparison Test (Priority: P1)

A musician wants to compare how different AI models interpret the same musical prompt to determine which model produces better rhythmic patterns. They run a comparison test with a specific prompt and review the resulting patterns from each model.

**Why this priority**: This is the core value proposition - without the ability to run comparisons, the feature has no purpose.

**Independent Test**: Can be fully tested by running a single prompt against two models and receiving saved results that can be played back and compared.

**Acceptance Scenarios**:

1. **Given** the user is in interplay, **When** they run a comparison with a prompt like "create a funky bass line", **Then** the system generates patterns from multiple configured models and saves each result.

2. **Given** a comparison is running, **When** each model completes generation, **Then** the user sees progress feedback indicating which model is processing.

3. **Given** a comparison completes, **When** the results are saved, **Then** each result includes the model name, prompt, generated commands, and resulting pattern state.

---

### User Story 2 - Review Saved Comparisons (Priority: P2)

A musician wants to review previously saved comparison results to evaluate patterns at their leisure, load specific patterns for playback, and make informed decisions about which model to use.

**Why this priority**: Without the ability to review saved results, comparisons would need to be evaluated immediately, limiting usefulness.

**Independent Test**: Can be tested by listing saved comparisons, viewing details of a specific comparison, and loading a pattern from a comparison result for playback.

**Acceptance Scenarios**:

1. **Given** comparisons have been saved, **When** the user lists comparisons, **Then** they see a list of all saved comparisons with dates, prompts, and models tested.

2. **Given** a saved comparison exists, **When** the user views its details, **Then** they see the prompt, each model's generated commands, and the resulting pattern for each model.

3. **Given** a saved comparison with multiple model results, **When** the user loads a specific model's result, **Then** that pattern becomes the active pattern for playback.

---

### User Story 3 - Select Model for Session (Priority: P3)

A musician wants to quickly switch which AI model is used for their current session without modifying configuration files, enabling rapid iteration when experimenting with different models.

**Why this priority**: Improves workflow efficiency but comparison and review features provide value even with manual model switching.

**Independent Test**: Can be tested by switching models via command or flag and verifying subsequent AI interactions use the selected model.

**Acceptance Scenarios**:

1. **Given** the user starts interplay, **When** they provide a model selection flag, **Then** all AI interactions use the specified model.

2. **Given** the user is in an AI session, **When** they switch models via command, **Then** subsequent AI interactions use the newly selected model.

3. **Given** multiple models are available, **When** the user requests to list available models, **Then** they see all configured models with their identifiers.

---

### User Story 4 - Rate Comparison Results (Priority: P4)

A musician wants to rate the results from each model after listening to them, building a record of subjective quality assessments that helps identify patterns in model performance over time.

**Why this priority**: Adds evaluation workflow but comparisons are useful even without formal ratings.

**Independent Test**: Can be tested by rating a model's output in a saved comparison and verifying the rating persists.

**Acceptance Scenarios**:

1. **Given** a saved comparison with results, **When** the user rates a model's output on defined criteria, **Then** the ratings are saved with the comparison.

2. **Given** multiple rated comparisons exist, **When** the user views comparison history, **Then** they can see aggregate ratings per model.

---

### User Story 5 - Blind Evaluation Mode (Priority: P3)

A musician wants to evaluate model outputs without knowing which model produced each pattern, eliminating bias and ensuring objective assessment of rhythmic quality.

**Why this priority**: Critical for unbiased evaluation - knowing the model name can influence perception of quality. Same priority as model selection since both enhance evaluation accuracy.

**Independent Test**: Can be tested by entering blind mode, listening to anonymized patterns, rating them, and verifying the model identities are revealed only after rating.

**Acceptance Scenarios**:

1. **Given** a saved comparison with multiple model results, **When** the user enters blind evaluation mode, **Then** results are presented with anonymous labels (A, B, C...) instead of model names.

2. **Given** blind mode is active, **When** the user loads pattern "A" for playback, **Then** the pattern plays without revealing which model generated it.

3. **Given** blind mode is active, **When** the user rates all anonymized patterns, **Then** the system reveals the mapping between labels and model names.

4. **Given** blind mode ratings are complete, **When** the reveal occurs, **Then** the ratings are saved with the correct model identifiers.

---

### Edge Cases

- What happens when a model API is unavailable during comparison? System should continue with available models and note the failure.
- What happens when a model returns invalid/unparseable commands? System should save the raw response and mark it as failed parsing.
- How does the system handle comparison prompts that result in empty patterns? System should save the result and note that no commands were generated.
- What happens when the user cancels a comparison mid-execution? System should save partial results and mark the comparison as incomplete.
- What happens if the user exits blind mode before rating all patterns? System should discard the incomplete blind session without saving partial ratings.
- What happens if a comparison has only one model result? System should warn that blind mode is not useful with a single result but allow it.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support configuring multiple AI models for comparison (initially Anthropic models: Haiku, Sonnet, Opus).
- **FR-002**: System MUST execute the same prompt against multiple selected models and capture each model's output.
- **FR-003**: System MUST save comparison results to the `comparisons/` directory with unique identifiers.
- **FR-004**: System MUST store for each comparison: timestamp, prompt text, and per-model results.
- **FR-005**: System MUST store for each model result: model identifier, generated commands, resulting pattern state, execution success/failure status.
- **FR-006**: System MUST provide commands to list, view, load, and delete saved comparisons. Comparisons are retained indefinitely until manually deleted.
- **FR-007**: System MUST allow users to switch the active AI model via command-line flag at startup. Default model is Haiku when no flag is provided.
- **FR-008**: System MUST allow users to switch the active AI model via command during an AI session.
- **FR-009**: System MUST list available models when requested by the user.
- **FR-010**: System MUST allow users to rate model outputs on defined criteria (rhythmic interest, velocity dynamics, genre accuracy, overall quality).
- **FR-011**: System MUST persist ratings with comparison results.
- **FR-012**: System MUST be designed to support additional AI providers in the future (extensible architecture).
- **FR-013**: System MUST handle model API failures gracefully, continuing comparison with available models.
- **FR-014**: System MUST handle invalid model responses by saving raw output and marking as parse failure.
- **FR-015**: System MUST provide a blind evaluation mode that presents model results with randomized anonymous labels (A, B, C...).
- **FR-016**: System MUST hide model identities during blind evaluation until all patterns have been rated.
- **FR-017**: System MUST reveal the label-to-model mapping after blind evaluation is complete.
- **FR-018**: System MUST ensure the randomized label assignment is different for each blind evaluation session to prevent pattern memorization.

### Key Entities

- **Comparison**: A single comparison run; contains timestamp, prompt, and collection of model results.
- **ModelResult**: Output from one model; contains model identifier, generated commands list, resulting pattern snapshot, success/failure status, optional error message.
- **ModelConfig**: Configuration for an available model; contains identifier, display name, provider type.
- **Rating**: User's subjective assessment; contains criteria scores (rhythmic interest, velocity dynamics, genre accuracy, overall) each rated 1-5.
- **BlindSession**: An active blind evaluation; contains comparison reference, randomized label-to-model mapping, and completion status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can run a comparison against 2+ models and receive saved results within 60 seconds per model.
- **SC-002**: Users can list and load any previously saved comparison within 2 seconds.
- **SC-003**: Users can switch between models in under 5 seconds without restarting the application.
- **SC-004**: 100% of comparison results are persisted and recoverable after application restart.
- **SC-005**: Users can identify which model produced a given pattern by viewing comparison details.
- **SC-006**: Users report increased confidence in model selection after using the comparison feature.

## Clarifications

### Session 2025-12-06

- Q: Which model should be used by default when no flag is provided? → A: Haiku (fastest/cheapest, explicit upgrade path for quality)
- Q: Where should comparison results be stored? → A: `comparisons/` directory (sibling to `patterns/`)
- Q: How long should comparisons be retained? → A: Keep indefinitely (manual cleanup via delete command)

## Assumptions

- Users have valid API keys for the AI providers they wish to compare.
- Network connectivity is available for API calls.
- The existing pattern storage format (JSON) is suitable for comparison result storage.
- Initial implementation focuses on Anthropic models; other providers will be added based on user demand.
- Rating criteria (rhythmic interest, velocity dynamics, genre accuracy, overall quality) are sufficient for initial evaluation; criteria may be refined based on user feedback.
