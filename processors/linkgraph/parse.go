package linkgraph

import (
	"bytes"
	"strings"
)

var skippedNamespaces = map[string]struct{}{
	"category": {}, "file": {}, "image": {}, "media": {}, "template": {},
	"help": {}, "wikipedia": {}, "module": {}, "draft": {}, "portal": {},
	"book": {}, "talk": {}, "user": {}, "special": {}, "mediawiki": {},
	"timedtext": {}, "education program": {}, "topic": {},
}

func ExtractPageLinkTargets(wikitext []byte) []string {
	seen := make(map[string]struct{})
	var res []string
	forEachWikilink(wikitext, func(inner []byte) {
		t := normalizeLinkTarget(inner)
		if t == "" || shouldSkipNamespace(t) {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		res = append(res, t)
	})
	return res
}

func forEachWikilink(text []byte, fn func(inner []byte)) {
	i := 0
	for i+1 < len(text) {
		if text[i] != '[' || text[i+1] != '[' {
			i++
			continue
		}
		start := i + 2
		depth := 1
		j := start
		for j+1 < len(text) && depth > 0 {
			if text[j] == '[' && text[j+1] == '[' {
				depth++
				j += 2
				continue
			}
			if text[j] == ']' && text[j+1] == ']' {
				depth--
				j += 2
				continue
			}
			j++
		}
		if depth != 0 {
			i = start
			continue
		}
		innerEnd := j - 2
		if innerEnd >= start {
			fn(text[start:innerEnd])
		}
		i = j
	}
}

func normalizeLinkTarget(inner []byte) string {
	if len(inner) == 0 {
		return ""
	}
	pipe := bytes.IndexByte(inner, '|')
	if pipe >= 0 {
		inner = inner[:pipe]
	}
	hash := bytes.IndexByte(inner, '#')
	if hash >= 0 {
		inner = inner[:hash]
	}
	s := string(bytes.TrimSpace(inner))
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.TrimSpace(s)
	return s
}

func shouldSkipNamespace(title string) bool {
	colon := strings.IndexByte(title, ':')
	if colon < 0 {
		return false
	}
	ns := strings.ToLower(strings.TrimSpace(title[:colon]))
	if ns == "" {
		return false
	}
	_, skip := skippedNamespaces[ns]
	return skip
}
