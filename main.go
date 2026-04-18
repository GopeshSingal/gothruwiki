package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"gothruwiki/internal/config"
	"gothruwiki/internal/processor"
	"gothruwiki/internal/wiki"
)

func main() {
	cfgPath, cfgExplicit := config.PathFromArgs()
	cfg, err := config.Load(cfgPath, cfgExplicit)
	if err != nil {
		log.Fatal(err)
	}

	flag.String("config", cfgPath, "JSON config file path (defaults for all flags; CLI overrides)")
	bz2File := flag.String("file", cfg.File, "Path to the enwiki .bz2 file")
	numWorkers := flag.Int("workers", cfg.Workers, "Number of workers")
	outFile := flag.String("out", cfg.Out, "Output file")
	checkpointFile := flag.String("checkpoint", cfg.Checkpoint, "File to save/load progress")
	checkpointEvery := flag.Int("checkpoint-every", cfg.CheckpointEvery, "Checkpoint every N stream pages (0 disables periodic checkpoints)")
	checkpointOnInterrupt := flag.Bool("checkpoint-on-interrupt", cfg.CheckpointOnInterrupt, "Save checkpoint when SIGINT or SIGTERM is received")
	resume := flag.Bool("resume", cfg.Resume, "Resume from the last checkpoint")
	procName := flag.String("processor", cfg.Processor, "Processor implementation name")
	decompress := flag.String("decompress", cfg.Decompress, "Decompression: auto (lbzip2, else stdlib), stdlib, lbzip2 (default: auto)")
	mainNamespaceOnly := flag.Bool("main-namespace-only", cfg.MainNamespaceOnly, "Process only main-namespace pages (ns=0); checkpoints still index stream page order")
	flag.Parse()

	if *bz2File == "" {
		log.Fatal("Error: -file is required.")
	}

	pf, ok := processorRegistry[*procName]
	if !ok {
		log.Fatalf("Unknown -processor %q; known: %s", *procName, strings.Join(knownProcessors(), ", "))
	}

	runner, err := pf().NewRunner(processor.Config{
		OutPath:        *outFile,
		CheckpointPath: *checkpointFile,
		Resume:         *resume,
	})
	if err != nil {
		log.Fatal(err)
	}

	startCount := 0
	if *resume {
		log.Printf("Attempting to resume from %s...", *checkpointFile)
		sc, err := runner.LoadCheckpoint()
		if err != nil {
			log.Printf("No checkpoint found or error loading: %v. Starting from scratch.", err)
		} else {
			startCount = sc
			log.Printf("Resuming from page %d...", startCount)
		}
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	mode := wiki.DecompressMode(strings.ToLower(strings.TrimSpace(*decompress)))
	decomp, decompLabel, err := wiki.OpenBZ2(*bz2File, mode)
	if err != nil {
		log.Fatalf("Open dump: %v", err)
	}
	defer decomp.Close()
	log.Printf("Decompressor: %s", decompLabel)
	if *mainNamespaceOnly {
		log.Printf("Filtering: main namespace only (ns=%d)", wiki.MainNamespace)
	}
	if *checkpointEvery > 0 {
		log.Printf("Periodic checkpoint every %d pages", *checkpointEvery)
	} else if *checkpointOnInterrupt {
		log.Printf("Periodic checkpoints disabled (SIGINT/SIGTERM will save checkpoint)")
	} else {
		log.Printf("Periodic checkpoints disabled (interrupt exits without checkpoint)")
	}

	safeReader := &wiki.SanitizerReader{Reader: decomp}

	jobs, wait := runner.StartWorkers(*numWorkers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parserErr := make(chan error, 1)
	var finalPageCount int

	go func() {
		pc, err := wiki.ParseDump(ctx, safeReader, wiki.StreamConfig{
			Jobs:              jobs,
			StartPageIndex:    startCount,
			MainNamespaceOnly: *mainNamespaceOnly,
			OnPage: func(pageCount int) {
				if pageCount%100000 == 0 {
					if pageCount < startCount {
						log.Printf("Fast-skipping: %d/%d pages...", pageCount, startCount)
					} else {
						msg := "Progress: " + strconv.Itoa(pageCount) + " pages processed..."
						if suf := runner.ProgressLogSuffix(pageCount); suf != "" {
							msg += " | " + suf
						}
						log.Println(msg)
					}
				}
				if *checkpointEvery > 0 && pageCount > 0 && pageCount%*checkpointEvery == 0 && pageCount > startCount {
					if err := runner.Checkpoint(pageCount); err != nil {
						log.Printf("Checkpoint error: %v", err)
					}
				}
			},
		})
		finalPageCount = pc
		parserErr <- err
	}()

	select {
	case err := <-parserErr:
		if err != nil && err != io.EOF && !errors.Is(err, context.Canceled) {
			log.Printf("Parser error: %v", err)
		}
	case <-stopChan:
		if *checkpointOnInterrupt {
			log.Println("\nInterrupt received! Saving progress before exiting...")
		} else {
			log.Println("\nInterrupt received! Exiting without saving checkpoint...")
		}
		cancel()
		_ = decomp.Close()
		err := <-parserErr
		if err != nil && err != io.EOF && !errors.Is(err, context.Canceled) {
			log.Printf("Parser error after cancel: %v", err)
		}
		close(jobs)
		wait()
		if *checkpointOnInterrupt {
			if err := runner.OnInterrupt(finalPageCount); err != nil {
				log.Printf("Checkpoint error: %v", err)
			}
			log.Println("Checkpoint saved. Exiting.")
		} else {
			log.Println("Exiting.")
		}
		os.Exit(0)
	}

	close(jobs)
	wait()
	if err := runner.Finalize(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Done! Final count: %d pages.", finalPageCount)
}
