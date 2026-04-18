package linkgraph

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
)

func adjacencyToSortedMap(edges map[string]map[string]struct{}) map[string][]string {
	out := make(map[string][]string, len(edges))
	for src, tgts := range edges {
		list := make([]string, 0, len(tgts))
		for t := range tgts {
			list = append(list, t)
		}
		sort.Strings(list)
		out[src] = list
	}
	return out
}

// TODO:
// Honestly, I'm pretty sure that, since the resultant graph is essentially a sparse matrix,
// there must be a better way to store this data. Perhaps if we also use an index file,
// though I am not sure if the page_id information is available online.
func SaveAdjacency(edges map[string]map[string]struct{}, outFile string) error {
	m := adjacencyToSortedMap(edges)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return err
		}
		valJSON, err := json.Marshal(m[k])
		if err != nil {
			return err
		}
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(valJSON)
	}
	buf.WriteByte('}')
	buf.WriteByte('\n')

	tmp := outFile + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, outFile)
}
