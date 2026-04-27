package pipeline

// Builder supports composable pipeline construction.
type Builder struct {
	stages []Stage
}

func NewBuilder() *Builder {
	return &Builder{
		stages: make([]Stage, 0),
	}
}

func (b *Builder) AddStage(stage Stage) *Builder {
	b.stages = append(b.stages, stage)
	return b
}

func (b *Builder) Build() *Pipeline {
	stages := make([]Stage, len(b.stages))
	copy(stages, b.stages)
	return New(stages...)
}
