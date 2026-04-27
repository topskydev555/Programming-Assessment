# Task 3 - Rate-Limited API Gateway Client

This task contains two parts:

- `go-client`: composable HTTP client wrapper with independent cross-cutting layers.
- `mobile-app`: React Native hook stack with dedupe, optimistic TTL cache, retry state, and cancellation.

## Go side

Path: `go-client/client`

Layers:

- `WithRateLimit`: blocks each request by calling a limiter abstraction.
- `WithRetry`: retries transient failures using exponential backoff.
- `WithResponseCache`: caches successful GET responses with TTL.
- `WithLogging`: records request and response lifecycle.

Composition uses `Chain` and `Client` in `client.go`, so each layer is swappable and testable in isolation.

Run tests:

```bash
cd "Task 3 Full-Stack - Rate-Limited API Gateway Client/go-client"
go test ./...
```

## React Native side

Path: `mobile-app/src`

Core logic:

- `ApiRequestManager`: pure request orchestration layer (dedupe/cache/retry/cancel).
- `useApiClient`: hook wrapper translating manager events into UI state.
- `ApiClientDemoScreen`: demo screen showing state transitions.

Run tests:

```bash
cd "Task 3 Full-Stack - Rate-Limited API Gateway Client/mobile-app"
npm test
```
