package stages

import (
	"context"
	"errors"
	"strings"
	"testing"

	"task1/pipeline"
)

func TestTransformStage_Process(t *testing.T) {
	stage := NewTransformStage("name", func(value any) (any, error) {
		asString, ok := value.(string)
		if !ok {
			return nil, errors.New("not a string")
		}
		return strings.ToUpper(asString), nil
	})

	if err := stage.Setup(context.Background()); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	record := pipeline.Record{
		ID: "1",
		Fields: map[string]any{
			"name": "alex",
		},
	}

	got, err := stage.Process(context.Background(), record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Fields["name"] != "ALEX" {
		t.Fatalf("expected transformed value ALEX, got %v", got.Fields["name"])
	}

	if record.Fields["name"] != "alex" {
		t.Fatalf("input record should stay unchanged")
	}
}

func TestTransformStage_SetupValidation(t *testing.T) {
	noField := NewTransformStage("", func(value any) (any, error) { return value, nil })
	if err := noField.Setup(context.Background()); err == nil {
		t.Fatalf("expected empty field setup error")
	}

	noTransformer := NewTransformStage("name", nil)
	if err := noTransformer.Setup(context.Background()); err == nil {
		t.Fatalf("expected nil transformer setup error")
	}
}
