package wordcount

import (
	"strings"
	"unsafe"
)

func AddCountsFromText(text []byte, into map[string]int) {
	wordStart := -1
	for i := 0; i < len(text); i++ {
		c := text[i]
		if isASCIILetter(c) {
			if wordStart < 0 {
				wordStart = i
			}
			continue
		}
		if wordStart >= 0 {
			addWord(text, wordStart, i, into)
			wordStart = -1
		}
	}
	if wordStart >= 0 {
		addWord(text, wordStart, len(text), into)
	}
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func addWord(text []byte, start, end int, into map[string]int) {
	n := end - start
	if n == 0 {
		return
	}
	const stackMax = 256
	var w string
	if n <= stackMax {
		var buf [stackMax]byte
		slc := buf[:n]
		for j := 0; j < n; j++ {
			c := text[start+j]
			if c >= 'A' && c <= 'Z' {
				slc[j] = c + ('a' - 'A')
			} else {
				slc[j] = c
			}
		}
		w = string(slc)
	} else {
		var b strings.Builder
		b.Grow(n)
		for j := 0; j < n; j++ {
			c := text[start+j]
			if c >= 'A' && c <= 'Z' {
				b.WriteByte(c + ('a' - 'A'))
			} else {
				b.WriteByte(c)
			}
		}
		w = b.String()
	}
	if len(w) > 1 || w == "a" || w == "i" {
		if isJunkWord(w) {
			return
		}
		into[w]++
	}
}

// CountWords applies the same tokenization rules as the streaming workers.
func CountWords(text string) map[string]int {
	m := make(map[string]int)
	AddCountsFromText(unsafe.Slice(unsafe.StringData(text), len(text)), m)
	return m
}
