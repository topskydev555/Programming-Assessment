package stages

import (
	"context"
	"errors"
	"fmt"

	"task1/pipeline"
)

type ValidationStage struct {
	requiredFields []string
}

func NewValidationStage(requiredFields ...string) *ValidationStage {
	return &ValidationStage{requiredFields: requiredFields}
}

func (s *ValidationStage) Name() string {
	return "validation"
}

func (s *ValidationStage) Setup(_ context.Context) error {
	if len(s.requiredFields) == 0 {
		return errors.New("validation requires at least one field")
	}
	return nil
}

func (s *ValidationStage) Process(_ context.Context, record pipeline.Record) (pipeline.Record, error) {
	for _, field := range s.requiredFields {
		value, ok := record.Fields[field]
		if !ok || value == nil {
			return pipeline.Record{}, fmt.Errorf("missing required field: %s", field)
		}

		if asString, isString := value.(string); isString && asString == "" {
			return pipeline.Record{}, fmt.Errorf("required field empty: %s", field)
		}
	}

	return record, nil
}

func (s *ValidationStage) Teardown(_ context.Context) error {
	return nil
}
