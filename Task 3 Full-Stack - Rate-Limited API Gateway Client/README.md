# Task 3 - Full-Stack Rate-Limited API Gateway Client

This task is split into two independent modules:

- `go-client`: composable Go HTTP client wrapper with cross-cutting layers
- `mobile-app`: Expo React Native app with `useApiClient` hook and demo UI

## Go Side

### Features

- Base `Doer` abstraction (`Do(req)`) for layer composition
- Rate limiting layer (token bucket)
- Retry layer with exponential backoff
- TTL response cache layer for GET requests
- Request/response logging layer
- Tests for each layer in isolation and composed behavior

### Run tests

```bash
cd go-client
go test ./...
```

## React Native Side

### Features

- `useApiClient` hook with:
  - request deduplication
  - optimistic cache (TTL)
  - auto retry with exponential backoff
  - cancellation support on unmount
- Demo screen with visible state transitions (`idle`, `loading`, `success`, `error`, `cancelled`)
- Request manager extracted into pure TypeScript class for unit testing

### Run app and tests

```bash
cd mobile-app
npm install
npm run test
npm start
```

## Design notes

- Each cross-cutting concern remains independent and composable.
- Both modules keep framework-specific code at the edges.
- Core behavior is testable without network calls or UI runtime.
