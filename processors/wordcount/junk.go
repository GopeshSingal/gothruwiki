package wordcount

var junkWords = map[string]struct{}{
	"access": {}, "archive": {}, "author": {}, "authors": {}, "cite": {},
	"date": {}, "editor": {}, "editors": {}, "file": {},
	"html": {}, "http": {}, "https": {}, "isbn": {}, "issn": {}, "journal": {},
	"last": {}, "name": {}, "page": {}, "pages": {}, "pmc": {}, "pmid": {},
	"publisher": {}, "ref": {}, "status": {}, "title": {}, "url": {},
	"website": {}, "work": {}, "arxiv": {}, "asin": {}, "bibcode": {}, "doi": {}, "jstor": {},
	"lccn": {}, "oclc": {}, "com": {}, "edu": {}, "gov": {}, "int": {}, "mil": {}, "net": {},
	"org": {}, "web": {}, "www": {}, "align": {}, "bgcolor": {}, "caption": {}, "category": {}, "center": {},
	"colspan": {}, "jpeg": {}, "jpg": {}, "pdf": {},
	"png": {}, "rowspan": {}, "style": {}, "svg": {}, "th": {}, "thumb": {},
	"thumbnail": {}, "wikipedia": {}, "php": {}, "utc": {}, "span": {}, "px": {}, "nbsp": {},
	"de": {}, "en": {}, "es": {}, "fr": {}, "it": {}, "ja": {}, "ko": {}, "pt": {}, "ru": {}, "zh": {},
}

func isJunkWord(w string) bool {
	_, ok := junkWords[w]
	return ok
}
