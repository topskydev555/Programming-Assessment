package stages

import (
	"context"
	"testing"

	"task1/pipeline"
)

func TestValidationStage_Process(t *testing.T) {
	stage := NewValidationStage("email", "name")
	if err := stage.Setup(context.Background()); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	valid := pipeline.Record{
		ID: "1",
		Fields: map[string]any{
			"email": "x@y.com",
			"name":  "alex",
		},
	}

	if _, err := stage.Process(context.Background(), valid); err != nil {
		t.Fatalf("expected valid record, got error: %v", err)
	}

	invalid := pipeline.Record{
		ID: "2",
		Fields: map[string]any{
			"name": "alex",
		},
	}

	if _, err := stage.Process(context.Background(), invalid); err == nil {
		t.Fatalf("expected error for missing field")
	}
}

func TestValidationStage_SetupRequiresFields(t *testing.T) {
	stage := NewValidationStage()
	if err := stage.Setup(context.Background()); err == nil {
		t.Fatalf("expected setup error when no fields configured")
	}
}
