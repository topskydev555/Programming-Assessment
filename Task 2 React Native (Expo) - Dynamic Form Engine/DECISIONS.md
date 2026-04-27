# Task 2 Decisions

## Architecture choices

- Split the solution into a pure TypeScript core (`src/engine`) and a React Native rendering layer (`src/components`, `src/fields`) to keep business logic testable in isolation.
- Centralized schema and form contracts in `src/types/schema.ts` so validation, visibility, state transitions, and rendering all share one typed source of truth.
- Kept form state orchestration in a dedicated `FormController` class instead of scattering logic in UI components.
- Implemented UI integration through a thin hook (`useDynamicForm`) that adapts controller state to React without coupling core logic to React lifecycle concerns.
- Treated unsupported field types as safe runtime errors in the renderer (clear message, no crash) to improve resilience.

## Patterns used and why

- **Registry pattern** (`FieldRegistry`): supports runtime registration of custom field components without modifying renderer internals.
- **Strategy pattern** (validation rules + custom validators): each rule is evaluated independently and composed per field, making validation behavior easy to extend.
- **State machine style transitions** (`pristine -> dirty -> validating -> submitting -> success/error`): makes form behavior explicit and predictable.
- **Adapter pattern** (`useDynamicForm`): bridges controller API to React state updates while preserving loose coupling.

## Alternatives considered and rejected

- A single hook containing all state, validation, and rendering decisions: rejected because it mixes concerns and makes isolated testing harder.
- Per-field hardcoded validation functions inside input components: rejected because rules become duplicated and harder to compose from schema.
- Rendering-only schema handling without typed contracts: rejected because dynamic forms become brittle and error-prone as schema evolves.
- Immediate crash on unknown field types: rejected because it breaks the full form for one schema issue and harms debuggability.

## Error handling and reliability choices

- Validation skips hidden fields by design, preventing false errors for conditionally invisible inputs.
- Submit flow validates before network work, then moves to `error` or `success` based on actual outcome.
- Missing custom validator keys return explicit validation messages rather than failing silently.
- Unknown field types are surfaced as clear inline messages so the rest of the form remains usable.

## Testability approach

- Core behavior is tested with unit tests against `FormController` (`src/engine/formController.test.ts`) without React Native rendering.
- Tests focus on transitions and rules that often regress in dynamic forms:
  - pristine to dirty tracking
  - hidden-field validation behavior
  - custom validator lookup and execution
  - submit success state transition

## AI usage and review changes

- Used AI to speed up initial scaffolding for schema types, controller structure, renderer wiring, and baseline tests.
- Reviewed and changed generated output to keep the design loosely coupled, reduce unnecessary complexity, and keep comments simple.
- Removed non-essential noise (extra logs and over-explained comments) and tightened error handling paths to keep behavior explicit.
