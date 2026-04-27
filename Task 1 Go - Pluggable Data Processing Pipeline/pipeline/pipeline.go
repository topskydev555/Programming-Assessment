package pipeline

import (
	"context"
	"errors"
	"io"
)

// Pipeline orchestrates stage lifecycle and per-record processing.
type Pipeline struct {
	stages       []Stage
	deadLetters  []DeadLetter
	setupDoneIdx int
}

// New creates a pipeline with ordered stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{
		stages:       stages,
		setupDoneIdx: -1,
	}
}

// Run reads from source, processes through stages, and writes to sink.
func (p *Pipeline) Run(ctx context.Context, source Source, sink Sink) error {
	if err := p.setupStages(ctx); err != nil {
		_ = p.teardownStages(ctx)
		return err
	}
	defer func() {
		_ = p.teardownStages(context.Background())
	}()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		record, err := source.Next(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		processedRecord, ok := p.processRecord(ctx, record)
		if !ok {
			continue
		}

		if err := sink.Write(ctx, processedRecord); err != nil {
			p.deadLetters = append(p.deadLetters, DeadLetter{
				Record:    processedRecord.Clone(),
				StageName: "sink",
				Err:       err,
			})
		}
	}
}

// DeadLetters returns an immutable copy of collected dead letters.
func (p *Pipeline) DeadLetters() []DeadLetter {
	cloned := make([]DeadLetter, len(p.deadLetters))
	copy(cloned, p.deadLetters)
	return cloned
}

func (p *Pipeline) processRecord(ctx context.Context, record Record) (Record, bool) {
	current := record
	for _, stage := range p.stages {
		next, err := stage.Process(ctx, current)
		if err != nil {
			p.deadLetters = append(p.deadLetters, DeadLetter{
				Record:    current.Clone(),
				StageName: stage.Name(),
				Err:       err,
			})
			return Record{}, false
		}
		current = next
	}
	return current, true
}

func (p *Pipeline) setupStages(ctx context.Context) error {
	for idx, stage := range p.stages {
		if err := stage.Setup(ctx); err != nil {
			return err
		}
		p.setupDoneIdx = idx
	}
	return nil
}

func (p *Pipeline) teardownStages(ctx context.Context) error {
	var teardownErr error

	for idx := p.setupDoneIdx; idx >= 0; idx-- {
		if err := p.stages[idx].Teardown(ctx); err != nil {
			teardownErr = errors.Join(teardownErr, err)
		}
	}

	p.setupDoneIdx = -1
	return teardownErr
}
