package pipeline

import "context"

// Record is the unit processed by the pipeline.
type Record struct {
	ID     string
	Fields map[string]any
}

// Clone creates a copy to avoid accidental shared mutation.
func (r Record) Clone() Record {
	clonedFields := make(map[string]any, len(r.Fields))
	for k, v := range r.Fields {
		clonedFields[k] = v
	}

	return Record{
		ID:     r.ID,
		Fields: clonedFields,
	}
}

// Stage represents a processing unit with a full lifecycle.
type Stage interface {
	Name() string
	Setup(ctx context.Context) error
	Process(ctx context.Context, record Record) (Record, error)
	Teardown(ctx context.Context) error
}

// Source provides records to pipeline.
type Source interface {
	Next(ctx context.Context) (Record, error)
}

// Sink consumes processed records.
type Sink interface {
	Write(ctx context.Context, record Record) error
}

// DeadLetter stores per-record failures without halting the pipeline.
type DeadLetter struct {
	Record    Record
	StageName string
	Err       error
}
