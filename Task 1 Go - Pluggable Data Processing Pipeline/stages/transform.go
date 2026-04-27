package stages

import (
	"context"
	"errors"
	"fmt"

	"task1/pipeline"
)

type TransformerFunc func(value any) (any, error)

type TransformStage struct {
	field       string
	transformer TransformerFunc
}

func NewTransformStage(field string, transformer TransformerFunc) *TransformStage {
	return &TransformStage{
		field:       field,
		transformer: transformer,
	}
}

func (s *TransformStage) Name() string {
	return "transform"
}

func (s *TransformStage) Setup(_ context.Context) error {
	if s.field == "" {
		return errors.New("transform field cannot be empty")
	}
	if s.transformer == nil {
		return errors.New("transformer cannot be nil")
	}
	return nil
}

func (s *TransformStage) Process(_ context.Context, record pipeline.Record) (pipeline.Record, error) {
	value, ok := record.Fields[s.field]
	if !ok {
		return pipeline.Record{}, fmt.Errorf("field not found: %s", s.field)
	}

	transformed, err := s.transformer(value)
	if err != nil {
		return pipeline.Record{}, fmt.Errorf("transform failed for field %s: %w", s.field, err)
	}

	mutated := record.Clone()
	mutated.Fields[s.field] = transformed
	return mutated, nil
}

func (s *TransformStage) Teardown(_ context.Context) error {
	return nil
}
