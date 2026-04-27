# Architectural Decisions

## Task 3 - Full-Stack Rate-Limited API Gateway Client

### 1) Go client built with composable decorator-style layers

- I used a `Doer` abstraction and `Layer` function type so each cross-cutting concern wraps the next concern without direct coupling.
- This keeps rate limiting, retry, caching, and logging independently testable by replacing the layer below with a fake `Doer`.
- I selected this over a monolithic client implementation because a single implementation would make behavior interactions hard to isolate and harder to extend.

### 2) Retry and cache interaction explicitly controlled

- Caching stores only successful `2xx` GET responses.
- This avoids caching temporary failures (`5xx`) and breaking retry semantics.
- I rejected a "cache everything" option because it can freeze failure responses and hide eventual recovery.

### 3) Time and sleep behavior injected for deterministic tests

- The Go retry layer uses a `Sleeper` abstraction.
- The mobile request manager accepts `delayFn` and `nowFn`.
- This removes reliance on wall-clock timing in tests and keeps behavior deterministic.

### 4) React Native hook split from orchestration engine

- `ApiRequestManager` contains dedupe/cache/retry/cancel mechanics as framework-agnostic logic.
- `useApiClient` only maps manager events into UI state (`loading`, `retrying`, `fromCache`, attempts, error/data).
- I chose this split over embedding all logic in the hook to keep logic reusable and easier to unit test in isolation.

### 5) Cancellation model favors safety on unmount

- Hook cleanup calls `manager.cancelAll()` to abort inflight work.
- Aborted requests map to a typed cancellation error state.
- This prevents stale async completion from updating unmounted screens.

## Patterns Used

- **Decorator / middleware chain** for Go HTTP concerns.
- **Dependency inversion** via interfaces (`Doer`, `Limiter`, `Logger`) and injected timing abstractions.
- **Single responsibility** via dedicated layers and manager/hook separation.

## Alternatives Considered and Rejected

- **Single giant Go client**: rejected due to tight coupling and weak test isolation.
- **Hook-only request logic**: rejected because behavior testing would require full React rendering for all cases.
- **Global shared cache singleton**: rejected to avoid hidden shared state and cross-screen coupling.

## AI Usage and Review Notes

- AI was used to speed up scaffolding and initial layer/test implementations.
- I manually reviewed and adjusted interactions, including a real bug fix where cache incorrectly stored `500` responses and interfered with retry.
- I also cleaned conflicting legacy code fragments in mobile files to keep the final code coherent and testable.
