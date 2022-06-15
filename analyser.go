package crdbts

import (
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/token/stop"
)

var (
	stopWordsMap = make(analysis.TokenMap, 256)
)

func init() {
	_ = stopWordsMap.LoadBytes(en.EnglishStopWords)
}

type Analyser struct {
	analyzer *analysis.Analyzer
}

func NewAnalyser() *Analyser {
	return &Analyser{
		analyzer: &analysis.Analyzer{
			Tokenizer: NewUnicodeTokenizer(),
			TokenFilters: []analysis.TokenFilter{
				en.NewPossessiveFilter(),
				lowercase.NewLowerCaseFilter(),
				stop.NewStopTokensFilter(stopWordsMap),
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
	a.analyzer.Tokenizer.(*UnicodeTokenizer).Reset()
}
