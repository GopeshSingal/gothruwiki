package wordcount

import (
	"gothruwiki/internal/processor"
)

type Processor struct{}

var _ processor.Processor = Processor{}

func (Processor) Name() string {
	return "wordcount"
}

func (Processor) NewRunner(cfg processor.Config) (processor.Runner, error) {
	return newRunner(cfg), nil
}
