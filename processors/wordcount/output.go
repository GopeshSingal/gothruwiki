package wordcount

import (
	"encoding/json"
	"io"
	"os"
	"sort"
)

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type OutputFormat struct {
	Top100   []WordCount `json:"top_100"`
	Top1000  []WordCount `json:"top_1000"`
	Top10000 []WordCount `json:"top_10000"`
}

func WriteOutput(w io.Writer, globalCounts map[string]int) error {
	var sorted []WordCount
	for word, count := range globalCounts {
		sorted = append(sorted, WordCount{Word: word, Count: count})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Count > sorted[j].Count })
	var out OutputFormat
	n := len(sorted)
	var l100, l1000, l10000 int
	switch {
	case n < 1000:
		l100, l1000, l10000 = n, n, n
	case n < 10000:
		l100, l1000, l10000 = 100, 1000, n
	default:
		l100, l1000, l10000 = 100, 1000, 10000
	}
	out.Top100, out.Top1000, out.Top10000 = sorted[:l100], sorted[:l1000], sorted[:l10000]
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func SaveOutput(globalCounts map[string]int, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteOutput(f, globalCounts)
}
