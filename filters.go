package crdbts

import (
	unicode2 "unicode"

	"github.com/GetStream/kit/v3/xbytes"
	"github.com/GetStream/kit/v3/xstrings"
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/snowballstem"
	"github.com/blevesearch/snowballstem/english"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// UniqueTermFilter retains only the tokens which mark the first occurrence of
// a term. Tokens whose term appears in a preceding token are dropped.
type UniqueTermFilter struct{}

func NewUniqueTermFilter() *UniqueTermFilter {
	return &UniqueTermFilter{}
}

func (f *UniqueTermFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	encounteredTerms := make(map[string]struct{})
	j := 0
	for _, token := range input {
		term := string(token.Term)
		if _, ok := encounteredTerms[term]; ok {
			continue
		}
		encounteredTerms[term] = struct{}{}
		input[j] = token
		j++
	}
	return input[:j]
}

type UnaccentedFilter struct {
	t transform.Transformer
}

func NewUnaccentedFilter() *UnaccentedFilter {
	return &UnaccentedFilter{t: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode2.Mn)), norm.NFC)}
}

func (u *UnaccentedFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for i := 0; i < len(input); i++ {
		res, _, err := transform.Append(u.t, input[i].Term[:0], input[i].Term)
		if err == nil {
			input[i].Term = res
		}
	}
	return input
}

type EnglishStemmerFilter struct {
	env *snowballstem.Env
}

func NewEnglishStemmerFilter() *EnglishStemmerFilter {
	return &EnglishStemmerFilter{env: &snowballstem.Env{}}
}

func (s *EnglishStemmerFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for _, token := range input {
		s.env.SetCurrent(xbytes.UnsafeString(token.Term))
		english.Stem(s.env)
		token.Term = xstrings.UnsafeBytes(s.env.Current())
	}
	return input
}
