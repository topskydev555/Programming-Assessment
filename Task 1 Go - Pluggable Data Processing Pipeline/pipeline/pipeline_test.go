package pipeline

import (
	"context"
	"errors"
	"io"
	"testing"
)

type testStage struct {
	name          string
	setupFn       func(context.Context) error
	processFn     func(context.Context, Record) (Record, error)
	teardownFn    func(context.Context) error
	setupCalls    int
	processCalls  int
	teardownCalls int
}

func (s *testStage) Name() string { return s.name }

func (s *testStage) Setup(ctx context.Context) error {
	s.setupCalls++
	if s.setupFn != nil {
		return s.setupFn(ctx)
	}
	return nil
}

func (s *testStage) Process(ctx context.Context, record Record) (Record, error) {
	s.processCalls++
	if s.processFn != nil {
		return s.processFn(ctx, record)
	}
	return record, nil
}

func (s *testStage) Teardown(ctx context.Context) error {
	s.teardownCalls++
	if s.teardownFn != nil {
		return s.teardownFn(ctx)
	}
	return nil
}

type cancelSource struct{}

func (s *cancelSource) Next(ctx context.Context) (Record, error) {
	<-ctx.Done()
	return Record{}, ctx.Err()
}

type sequenceSource struct {
	records []Record
	idx     int
}

func (s *sequenceSource) Next(_ context.Context) (Record, error) {
	if s.idx >= len(s.records) {
		return Record{}, io.EOF
	}
	record := s.records[s.idx].Clone()
	s.idx++
	return record, nil
}

type collectingSink struct {
	records []Record
	failID  string
}

func (s *collectingSink) Write(_ context.Context, record Record) error {
	if s.failID != "" && record.ID == s.failID {
		return errors.New("sink failure")
	}
	s.records = append(s.records, record.Clone())
	return nil
}

func TestPipeline_Run_ContinuesOnRecordErrors(t *testing.T) {
	failingStage := &testStage{
		name: "failing",
		processFn: func(_ context.Context, record Record) (Record, error) {
			if record.ID == "bad" {
				return Record{}, errors.New("bad record")
			}
			return record, nil
		},
	}

	p := New(failingStage)
	source := &sequenceSource{
		records: []Record{
			{ID: "ok1", Fields: map[string]any{"v": 1}},
			{ID: "bad", Fields: map[string]any{"v": 2}},
			{ID: "ok2", Fields: map[string]any{"v": 3}},
		},
	}
	sink := &collectingSink{}

	if err := p.Run(context.Background(), source, sink); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(sink.records) != 2 {
		t.Fatalf("expected 2 records in sink, got %d", len(sink.records))
	}

	deadLetters := p.DeadLetters()
	if len(deadLetters) != 1 {
		t.Fatalf("expected 1 dead letter, got %d", len(deadLetters))
	}
	if deadLetters[0].StageName != "failing" {
		t.Fatalf("expected dead letter stage failing, got %s", deadLetters[0].StageName)
	}
}

func TestPipeline_Run_CollectsSinkErrors(t *testing.T) {
	p := New()
	source := &sequenceSource{
		records: []Record{
			{ID: "1", Fields: map[string]any{}},
			{ID: "2", Fields: map[string]any{}},
		},
	}
	sink := &collectingSink{failID: "2"}

	if err := p.Run(context.Background(), source, sink); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(sink.records) != 1 {
		t.Fatalf("expected 1 successful sink write, got %d", len(sink.records))
	}

	deadLetters := p.DeadLetters()
	if len(deadLetters) != 1 || deadLetters[0].StageName != "sink" {
		t.Fatalf("expected one sink dead letter")
	}
}

func TestPipeline_Run_PropagatesCancellation(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.Run(ctx, &cancelSource{}, &collectingSink{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestPipeline_Run_CallsLifecycleHooks(t *testing.T) {
	first := &testStage{name: "first"}
	second := &testStage{name: "second"}

	p := New(first, second)
	source := &sequenceSource{
		records: []Record{
			{ID: "1", Fields: map[string]any{}},
		},
	}
	sink := &collectingSink{}

	if err := p.Run(context.Background(), source, sink); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if first.setupCalls != 1 || second.setupCalls != 1 {
		t.Fatalf("expected one setup call per stage")
	}
	if first.processCalls != 1 || second.processCalls != 1 {
		t.Fatalf("expected one process call per stage")
	}
	if first.teardownCalls != 1 || second.teardownCalls != 1 {
		t.Fatalf("expected one teardown call per stage")
	}
}
