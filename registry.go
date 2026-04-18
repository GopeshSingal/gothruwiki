package main

import (
	"sort"

	"gothruwiki/internal/processor"
	"gothruwiki/processors/linkgraph"
	"gothruwiki/processors/wordcount"
)

var processorRegistry = map[string]func() processor.Processor{
	"wordcount": func() processor.Processor { return wordcount.Processor{} },
	"linkgraph": func() processor.Processor { return linkgraph.Processor{} },
}

func knownProcessors() []string {
	names := make([]string, 0, len(processorRegistry))
	for k := range processorRegistry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
