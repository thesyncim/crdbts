package crdbts

import (
	"unicode"
	"unicode/utf8"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/segment"
)

type UnicodeTokenizer struct {
	rall analysis.TokenStream
	rvx  []analysis.TokenStream
	rv   analysis.TokenStream
	ta   []analysis.Token
}

func NewUnicodeTokenizer() *UnicodeTokenizer {
	return &UnicodeTokenizer{
		rvx: make([]analysis.TokenStream, 0, 10),
		rv:  make(analysis.TokenStream, 0, 1),
		ta:  make([]analysis.Token, 1),
	}
}

func (rt *UnicodeTokenizer) Reset() {
	rt.rall = rt.rall[:0]
	rt.rvx = rt.rvx[:0]
	rt.rv = rt.rv[:0]
}

func (rt *UnicodeTokenizer) Tokenize(input []byte) analysis.TokenStream {
	rt.Reset()

	taNext := 0

	segmenter := segment.NewWordSegmenterDirect(input)
	start := 0
	pos := 1

	guessRemaining := func(end int) int {
		avgSegmentLen := end / (len(rt.rv) + 1)
		if avgSegmentLen < 1 {
			avgSegmentLen = 1
		}

		remainingLen := len(input) - end

		return remainingLen / avgSegmentLen
	}

	for segmenter.Segment() {
		segmentBytes := segmenter.Bytes()
		end := start + len(segmentBytes)
		if segmenter.Type() != segment.None || !shouldSkip(segmentBytes) {
			if taNext >= len(rt.ta) {
				remainingSegments := guessRemaining(end)
				if remainingSegments > 1000 {
					remainingSegments = 1000
				}
				if remainingSegments < 1 {
					remainingSegments = 1
				}

				rt.ta = make([]analysis.Token, remainingSegments)
				taNext = 0
			}

			token := &rt.ta[taNext]
			taNext++

			token.Term = segmentBytes
			token.Start = start
			token.End = end
			token.Position = pos
			token.Type = convertType(segmenter.Type())

			if len(rt.rv) >= cap(rt.rv) { // When rv is full, save it into rvx.
				rt.rvx = append(rt.rvx, rt.rv)

				rvCap := cap(rt.rv) * 2
				if rvCap > 256 {
					rvCap = 256
				}

				rt.rv = make(analysis.TokenStream, 0, rvCap) // Next rv cap is bigger.
			}

			rt.rv = append(rt.rv, token)
			pos++
		}
		start = end
	}

	if len(rt.rvx) > 0 {
		for _, r := range rt.rvx {
			rt.rall = append(rt.rall, r...)
		}
		return append(rt.rall, rt.rv...)
	}

	return rt.rv
}

func shouldSkip(bytes []byte) bool {
	if len(bytes) < 1 || len(bytes) > 4 {
		return true
	}
	r, n := utf8.DecodeRune(bytes)
	if n == 0 {
		return true
	}
	if unicode.IsSpace(r) {
		return true
	}
	if unicode.IsPunct(r) {
		return true
	}
	return false
}

func convertType(segmentWordType int) analysis.TokenType {
	switch segmentWordType {
	case segment.Ideo:
		return analysis.Ideographic
	case segment.Kana:
		return analysis.Ideographic
	case segment.Number:
		return analysis.Numeric
	}
	return analysis.AlphaNumeric
}
