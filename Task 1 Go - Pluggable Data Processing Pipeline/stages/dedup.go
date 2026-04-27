package stages

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"task1/pipeline"
)

type DedupStage struct {
	keyField string
	seen     map[string]struct{}
	mu       sync.Mutex
}

func NewDedupStage(keyField string) *DedupStage {
	return &DedupStage{keyField: keyField}
}

func (s *DedupStage) Name() string {
	return "dedup"
}

func (s *DedupStage) Setup(_ context.Context) error {
	if s.keyField == "" {
		return errors.New("dedup key field cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = make(map[string]struct{})
	return nil
}

func (s *DedupStage) Process(_ context.Context, record pipeline.Record) (pipeline.Record, error) {
	value, ok := record.Fields[s.keyField]
	if !ok {
		return pipeline.Record{}, fmt.Errorf("dedup key missing: %s", s.keyField)
	}

	key := fmt.Sprintf("%v", value)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.seen[key]; exists {
		return pipeline.Record{}, fmt.Errorf("duplicate key: %s", key)
	}

	s.seen[key] = struct{}{}
	return record, nil
}

func (s *DedupStage) Teardown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = nil
	return nil
}
