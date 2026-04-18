package wordcount

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"

	"gothruwiki/internal/processor"
	"gothruwiki/internal/wiki"
)

type checkpointData struct {
	PageCount  int
	WordCounts map[string]int
}

type sharedState struct {
	Counts map[string]int
	Mu     sync.Mutex
}

type runner struct {
	cfg   processor.Config
	state *sharedState
}

func newRunner(cfg processor.Config) *runner {
	return &runner{
		cfg: cfg,
		state: &sharedState{
			Counts: make(map[string]int),
		},
	}
}

func (r *runner) StartWorkers(numWorkers int) (chan wiki.Page, func()) {
	jobs := make(chan wiki.Page, numWorkers*8)
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go r.worker(jobs, &wg)
	}
	wait := func() {
		wg.Wait()
	}
	return jobs, wait
}

func (r *runner) worker(jobs <-chan wiki.Page, wg *sync.WaitGroup) {
	defer wg.Done()
	localMap := make(map[string]int)
	processed := 0
	for job := range jobs {
		AddCountsFromText(job.Text, localMap)
		processed++
		if processed%1000 == 0 {
			r.state.Mu.Lock()
			for k, v := range localMap {
				r.state.Counts[k] += v
			}
			r.state.Mu.Unlock()
			for k := range localMap {
				delete(localMap, k)
			}
		}
	}
	r.state.Mu.Lock()
	for k, v := range localMap {
		r.state.Counts[k] += v
	}
	r.state.Mu.Unlock()
}

func (r *runner) LoadCheckpoint() (int, error) {
	if !r.cfg.Resume {
		return 0, nil
	}
	var cp checkpointData
	f, err := os.Open(r.cfg.CheckpointPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&cp); err != nil {
		return 0, err
	}
	if cp.WordCounts == nil {
		cp.WordCounts = make(map[string]int)
	}
	r.state.Mu.Lock()
	r.state.Counts = cp.WordCounts
	r.state.Mu.Unlock()
	return cp.PageCount, nil
}

func (r *runner) ProgressLogSuffix(pageCount int) string {
	_ = pageCount
	r.state.Mu.Lock()
	n := len(r.state.Counts)
	r.state.Mu.Unlock()
	return fmt.Sprintf("unique_words=%d", n)
}

func (r *runner) Checkpoint(pageCount int) error {
	r.state.Mu.Lock()
	defer r.state.Mu.Unlock()

	tmp := r.cfg.CheckpointPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		log.Printf("Failed to create checkpoint: %v", err)
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(checkpointData{PageCount: pageCount, WordCounts: r.state.Counts})
	f.Close()
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, r.cfg.CheckpointPath); err != nil {
		return err
	}
	log.Printf("---> Checkpoint saved at page %d", pageCount)
	return nil
}

func (r *runner) Finalize() error {
	r.state.Mu.Lock()
	counts := r.state.Counts
	r.state.Mu.Unlock()
	return SaveOutput(counts, r.cfg.OutPath)
}

func (r *runner) OnInterrupt(pageCount int) error {
	return r.Checkpoint(pageCount)
}

var _ processor.Runner = (*runner)(nil)
