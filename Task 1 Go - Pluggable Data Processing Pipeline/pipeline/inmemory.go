package pipeline

import (
	"context"
	"io"
)

type InMemorySource struct {
	records []Record
	index   int
}

func NewInMemorySource(records []Record) *InMemorySource {
	cloned := make([]Record, len(records))
	for i := range records {
		cloned[i] = records[i].Clone()
	}
	return &InMemorySource{records: cloned}
}

func (s *InMemorySource) Next(_ context.Context) (Record, error) {
	if s.index >= len(s.records) {
		return Record{}, io.EOF
	}
	record := s.records[s.index].Clone()
	s.index++
	return record, nil
}

type InMemorySink struct {
	records []Record
}

func NewInMemorySink() *InMemorySink {
	return &InMemorySink{records: make([]Record, 0)}
}

func (s *InMemorySink) Write(_ context.Context, record Record) error {
	s.records = append(s.records, record.Clone())
	return nil
}

func (s *InMemorySink) Records() []Record {
	cloned := make([]Record, len(s.records))
	for i := range s.records {
		cloned[i] = s.records[i].Clone()
	}
	return cloned
}
