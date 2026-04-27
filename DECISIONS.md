# Assessment Decisions

## Key architecture choices

- Built Task 3 as two modules (`go-client` and `mobile-app`) to keep backend-style transport concerns and frontend state concerns isolated.
- Kept all cross-cutting concerns as composable wrappers instead of embedding logic into one large client implementation.
- Used interface-first design (`Doer` in Go, `ApiClient` in TypeScript) so each layer can be unit tested by mocking the layer below.

## Patterns used and why

- **Decorator / Middleware** (Go and TS): best fit for stacking rate limit, retry, cache, and logging without tight coupling.
- **Strategy** (`ShouldRetry`, `ApiClient`): allows behavior changes without rewiring orchestration.
- **Adapter** (`HTTPClient`, `FetchApiClient`): isolates third-party/runtime APIs (`http.Client`, `fetch`) behind project contracts.

## Alternatives considered and rejected

- Monolithic HTTP wrapper with all concerns in one method: rejected due to low testability and high change risk.
- Global singleton cache/retry manager in the mobile layer: rejected to avoid hidden mutable state across screens.
- State management library for Task 3 mobile demo: rejected as unnecessary for scope; hook + manager kept simpler and testable.

## Testing approach

- Go tests validate each layer independently and a composed pipeline path.
- TypeScript tests validate request manager behavior: dedupe, retry, and TTL cache.
- Mocks/stubs used for deterministic behavior and no external network dependency.

## AI usage and review changes

- Used AI to accelerate scaffolding of module structure, middleware contracts, and initial tests.
- Revised generated code to keep comments simple, avoid unnecessary logging, and enforce loose coupling.
- Added explicit cancellation handling, TTL boundaries, and predictable test doubles after review.
