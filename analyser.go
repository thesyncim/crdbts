package crdbts

import (
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/registry"
)

type Analyser struct {
	tokenizer *UnicodeTokenizer
	analyzer  *analysis.Analyzer
}

var cache = registry.NewCache()

func NewAnalyser() *Analyser {
	stopFilter, err := cache.TokenFilterNamed(en.StopName)
	if err != nil {
		panic(err)
	}

	tokenizer := NewUnicodeTokenizer()

	return &Analyser{
		tokenizer: tokenizer,
		analyzer: &analysis.Analyzer{
			Tokenizer: tokenizer,
			TokenFilters: []analysis.TokenFilter{
				en.NewPossessiveFilter(),
				lowercase.NewLowerCaseFilter(),
				stopFilter,
				NewUniqueTermFilter(),
				NewUnaccentedFilter(),
				NewEnglishStemmerFilter(),
			},
		}}
}

func (a *Analyser) Analyse(text string) analysis.TokenStream {
	return a.analyzer.Analyze([]byte(text))
}

func (a *Analyser) Reset() {
	a.tokenizer.Reset()
}
