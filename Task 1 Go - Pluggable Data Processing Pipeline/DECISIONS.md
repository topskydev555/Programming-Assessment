# Task 1 Decisions

## Architecture choices

- Used a small core (`pipeline`) package with `Stage`, `Source`, and `Sink` interfaces to keep processing logic loosely coupled.
- Kept pipeline orchestration separate from stage implementations (`stages`) so each stage is independently testable and reusable.
- Added a `Builder` API to compose stages in order and satisfy pluggable configuration without introducing heavy framework code.
- Modeled dead-letter handling as a first-class output (`DeadLetter`) with record snapshot, stage name, and error.

## Patterns used and why

- **Pipeline pattern**: natural fit for ordered record transformations.
- **Strategy pattern** (via `Stage` interface): enables swapping/adding stage behavior without changing orchestration code.
- **Builder pattern** (`NewBuilder().AddStage(...).Build()`): keeps stage wiring readable and extensible.

## Alternatives considered and rejected

- A single monolithic processor function with embedded `if/else` stage logic: rejected because it couples behavior and makes extension risky.
- Generic middleware chain with reflection-heavy config loading: rejected because it adds complexity without clear value for this scope.
- Hard-failing pipeline on first record error: rejected because requirement asks for per-record fault tolerance.

## Error handling and shutdown

- `Run` checks `context.Context` each loop, returning `context.Canceled` or deadline errors directly.
- Stage setup/teardown lifecycle is explicit; teardown runs for stages that were successfully set up.
- Record processing errors and sink write errors are captured in dead letters so the stream keeps moving.

## AI usage and review changes

- Used AI to accelerate initial scaffolding for interfaces, stage implementations, and tests.
- Reviewed and adjusted generated output to keep interfaces minimal, improve cloning boundaries, and tighten lifecycle semantics.
- Removed unnecessary logging and kept comments simple to reduce noise and keep intent clear.
