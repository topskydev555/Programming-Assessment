# Task 1 - Pluggable Data Processing Pipeline (Go)

This project implements a configurable record processing pipeline with pluggable stages, per-record fault tolerance, and stage lifecycle management.

## Features

- Pluggable stages via interfaces and builder composition
- Stage lifecycle hooks: `Setup`, `Process`, `Teardown`
- Per-record error handling with dead-letter collection
- Graceful shutdown through `context.Context` cancellation
- Isolated unit tests for stages and orchestration

## Project Structure

- `main.go`: runnable example pipeline
- `pipeline/`
  - `types.go`: core interfaces and models
  - `pipeline.go`: orchestration and dead-letter flow
  - `builder.go`: stage composition builder
  - `inmemory.go`: in-memory source and sink adapters
- `stages/`
  - `validation.go`: required field validation
  - `transform.go`: field transformation stage
  - `dedup.go`: deduplication stage
  - `*_test.go`: stage unit tests
- `pipeline/pipeline_test.go`: orchestration tests
- `DECISIONS.md`: architecture and design decisions

## Requirements

- Go 1.22+

## Run

```bash
go run .
```

The demo runs three records through:

1. Validation (`email`, `name` required)
2. Name transformation (uppercase)
3. Deduplication by `email`

Successful outputs go to sink; failed records go to dead letters with stage and error context.

## Test

```bash
go test ./...
```

## Extending the Pipeline

Add a new stage by implementing:

```go
type Stage interface {
    Name() string
    Setup(ctx context.Context) error
    Process(ctx context.Context, record Record) (Record, error)
    Teardown(ctx context.Context) error
}
```

Then compose it using:

```go
p := pipeline.NewBuilder().
    AddStage(stageA).
    AddStage(stageB).
    Build()
```

## Error Handling Model

- Stage processing errors do not halt the pipeline
- Sink write errors are also captured as dead letters
- Dead letters include:
  - record snapshot
  - failing stage name
  - error details

