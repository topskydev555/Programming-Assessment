package main

import (
	"context"
	"fmt"
	"strings"

	"task1/pipeline"
	"task1/stages"
)

func main() {
	records := []pipeline.Record{
		{
			ID: "1",
			Fields: map[string]any{
				"email": "alice@example.com",
				"name":  "alice",
			},
		},
		{
			ID: "2",
			Fields: map[string]any{
				"email": "alice@example.com",
				"name":  "alice duplicate",
			},
		},
		{
			ID: "3",
			Fields: map[string]any{
				"email": "",
				"name":  "missing email",
			},
		},
	}

	p := pipeline.NewBuilder().
		AddStage(stages.NewValidationStage("email", "name")).
		AddStage(stages.NewTransformStage("name", func(value any) (any, error) {
			asString, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("name must be string")
			}
			return strings.ToUpper(asString), nil
		})).
		AddStage(stages.NewDedupStage("email")).
		Build()

	source := pipeline.NewInMemorySource(records)
	sink := pipeline.NewInMemorySink()

	if err := p.Run(context.Background(), source, sink); err != nil {
		panic(err)
	}

	fmt.Println("processed records:")
	for _, record := range sink.Records() {
		fmt.Printf("- id=%s fields=%v\n", record.ID, record.Fields)
	}

	fmt.Println("dead letters:")
	for _, deadLetter := range p.DeadLetters() {
		fmt.Printf("- id=%s stage=%s err=%v\n", deadLetter.Record.ID, deadLetter.StageName, deadLetter.Err)
	}
}
