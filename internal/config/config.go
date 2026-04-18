package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const DefaultFilename = "gothruwiki.json"

func PathFromArgs() (path string, explicit bool) {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-config" || a == "--config":
			explicit = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], true
			}
			return DefaultFilename, true
		case strings.HasPrefix(a, "-config="):
			explicit = true
			return strings.TrimPrefix(a, "-config="), true
		case strings.HasPrefix(a, "--config="):
			explicit = true
			return strings.TrimPrefix(a, "--config="), true
		}
	}
	return DefaultFilename, false
}

type App struct {
	File                    string `json:"file"`
	Workers                 int    `json:"workers"`
	Out                     string `json:"out"`
	Checkpoint              string `json:"checkpoint"`
	CheckpointEvery         int    `json:"checkpoint_every"`
	CheckpointOnInterrupt   bool   `json:"checkpoint_on_interrupt"`
	Resume                  bool   `json:"resume"`
	Processor               string `json:"processor"`
	Decompress              string `json:"decompress"`
	MainNamespaceOnly       bool   `json:"main_namespace_only"`
}

func Defaults() App {
	return App{
		Workers:         8,
		Checkpoint:      "wiki_progress.gob",
		CheckpointEvery: 1_000_000,
		Resume:          false,
		Decompress:      "auto",
	}
}

func Load(path string, mustExist bool) (App, error) {
	d := Defaults()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if mustExist {
				return d, fmt.Errorf("config file not found: %s", path)
			}
			return d, nil
		}
		return d, err
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return Defaults(), fmt.Errorf("parse config %s: %w", path, err)
	}
	d.normalize()
	return d, nil
}

func (d *App) normalize() {
	def := Defaults()
	if d.Workers <= 0 {
		d.Workers = def.Workers
	}
	if strings.TrimSpace(d.Out) == "" {
		d.Out = def.Out
	}
	if strings.TrimSpace(d.Checkpoint) == "" {
		d.Checkpoint = def.Checkpoint
	}
	if strings.TrimSpace(d.Processor) == "" {
		d.Processor = def.Processor
	}
	if strings.TrimSpace(d.Decompress) == "" {
		d.Decompress = def.Decompress
	}
}
