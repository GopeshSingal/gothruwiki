package processor

import "gothruwiki/internal/wiki"

type Config struct {
	OutPath        string
	CheckpointPath string
	Resume         bool
}

type Processor interface {
	Name() string
	NewRunner(cfg Config) (Runner, error)
}

type Runner interface {
	StartWorkers(numWorkers int) (jobs chan wiki.Page, wait func())
	LoadCheckpoint() (startPageIndex int, err error)
	ProgressLogSuffix(pageCount int) string
	Checkpoint(pageCount int) error
	Finalize() error
	OnInterrupt(pageCount int) error
}
