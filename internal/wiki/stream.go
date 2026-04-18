package wiki

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/lestrrat-go/helium"
	"github.com/lestrrat-go/helium/sax"
)

const MainNamespace = 0

type StreamConfig struct {
	Jobs              chan<- Page
	StartPageIndex    int
	MainNamespaceOnly bool
	OnPage            func(pageCount int)
}

func ParseDump(ctx context.Context, r io.Reader, cfg StreamConfig) (pageCount int, err error) {
	handler := sax.New()
	var inText, inNS, inTitle bool
	var textBuf, nsBuf, titleBuf bytes.Buffer
	pageNS := -1

	handler.SetOnStartElementNS(sax.StartElementNSFunc(func(_ context.Context, localname, _, _ string, _ []sax.Namespace, _ []sax.Attribute) error {
		switch localname {
		case "page":
			pageNS = -1
			titleBuf.Reset()
		case "title":
			inTitle = true
			titleBuf.Reset()
		case "ns":
			inNS = true
			nsBuf.Reset()
		case "text":
			inText = true
			textBuf.Reset()
		}
		return nil
	}))

	handler.SetOnCharacters(sax.CharactersFunc(func(_ context.Context, ch []byte) error {
		switch {
		case inTitle:
			titleBuf.Write(ch)
		case inNS:
			nsBuf.Write(ch)
		case inText:
			textBuf.Write(ch)
		}
		return nil
	}))

	handler.SetOnEndElementNS(sax.EndElementNSFunc(func(_ context.Context, localname, _, _ string) error {
		switch localname {
		case "title":
			inTitle = false
		case "ns":
			inNS = false
			n, err := strconv.Atoi(strings.TrimSpace(nsBuf.String()))
			if err == nil {
				pageNS = n
			} else {
				pageNS = -1
			}
		case "text":
			inText = false
			if pageCount >= cfg.StartPageIndex &&
				(!cfg.MainNamespaceOnly || pageNS == MainNamespace) {
				cfg.Jobs <- Page{
					Title: strings.TrimSpace(titleBuf.String()),
					Text:  bytes.Clone(textBuf.Bytes()),
				}
			}
		case "page":
			pageCount++
			if cfg.OnPage != nil {
				cfg.OnPage(pageCount)
			}
		}
		return nil
	}))

	parser := helium.NewParser().SAXHandler(handler)
	if ctx == nil {
		ctx = context.Background()
	}
	_, err = parser.ParseReader(ctx, r)
	return pageCount, err
}
