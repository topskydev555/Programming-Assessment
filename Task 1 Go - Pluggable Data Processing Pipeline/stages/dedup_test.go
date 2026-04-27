package stages

import (
	"context"
	"testing"

	"task1/pipeline"
)

func TestDedupStage_Process(t *testing.T) {
	stage := NewDedupStage("email")
	if err := stage.Setup(context.Background()); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer func() { _ = stage.Teardown(context.Background()) }()

	first := pipeline.Record{
		ID: "1",
		Fields: map[string]any{
			"email": "x@y.com",
		},
	}

	second := pipeline.Record{
		ID: "2",
		Fields: map[string]any{
			"email": "x@y.com",
		},
	}

	if _, err := stage.Process(context.Background(), first); err != nil {
		t.Fatalf("first record should pass: %v", err)
	}

	if _, err := stage.Process(context.Background(), second); err == nil {
		t.Fatalf("expected duplicate error")
	}
}

func TestDedupStage_SetupValidation(t *testing.T) {
	stage := NewDedupStage("")
	if err := stage.Setup(context.Background()); err == nil {
		t.Fatalf("expected setup error when key empty")
	}
}
