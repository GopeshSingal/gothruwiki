package linkgraph

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
	PageCount int
	Adj       map[string][]string
}

type sharedState struct {
	Edges map[string]map[string]struct{}
	Mu    sync.Mutex
}

type runner struct {
	cfg   processor.Config
	state *sharedState
}

func newRunner(cfg processor.Config) *runner {
	return &runner{
		cfg: cfg,
		state: &sharedState{
			Edges: make(map[string]map[string]struct{}),
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
	local := make(map[string]map[string]struct{})
	processed := 0
	for job := range jobs {
		src := job.Title
		if src == "" {
			continue
		}
		for _, tgt := range ExtractPageLinkTargets(job.Text) {
			if tgt == "" || tgt == src {
				continue
			}
			m := local[src]
			if m == nil {
				m = make(map[string]struct{})
				local[src] = m
			}
			m[tgt] = struct{}{}
		}
		processed++
		if processed%1000 == 0 {
			r.mergeIn(local)
			for k := range local {
				delete(local, k)
			}
		}
	}
	r.mergeIn(local)
}

func (r *runner) mergeIn(local map[string]map[string]struct{}) {
	if len(local) == 0 {
		return
	}
	r.state.Mu.Lock()
	defer r.state.Mu.Unlock()
	for src, tgts := range local {
		dst := r.state.Edges[src]
		if dst == nil {
			dst = make(map[string]struct{})
			r.state.Edges[src] = dst
		}
		for t := range tgts {
			dst[t] = struct{}{}
		}
	}
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
	if cp.Adj == nil {
		cp.Adj = make(map[string][]string)
	}
	edges := make(map[string]map[string]struct{}, len(cp.Adj))
	for src, list := range cp.Adj {
		m := make(map[string]struct{}, len(list))
		for _, t := range list {
			m[t] = struct{}{}
		}
		edges[src] = m
	}
	r.state.Mu.Lock()
	r.state.Edges = edges
	r.state.Mu.Unlock()
	return cp.PageCount, nil
}

func (r *runner) ProgressLogSuffix(pageCount int) string {
	_ = pageCount
	r.state.Mu.Lock()
	n := len(r.state.Edges)
	r.state.Mu.Unlock()
	return fmt.Sprintf("sources=%d", n)
}

func (r *runner) Checkpoint(pageCount int) error {
	r.state.Mu.Lock()
	adj := adjacencyToSortedMap(r.state.Edges)
	r.state.Mu.Unlock()

	tmp := r.cfg.CheckpointPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		log.Printf("Failed to create checkpoint: %v", err)
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(checkpointData{PageCount: pageCount, Adj: adj})
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
	edges := r.state.Edges
	r.state.Mu.Unlock()
	return SaveAdjacency(edges, r.cfg.OutPath)
}

func (r *runner) OnInterrupt(pageCount int) error {
	return r.Checkpoint(pageCount)
}

var _ processor.Runner = (*runner)(nil)
