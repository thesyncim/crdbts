package crdbts

import (
	"context"

	"github.com/GetStream/kit/v3/xbytes"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func Search(db orm.DB, term string) ([]*MessageSearch, error) {
	var messages []*MessageSearch
	terms := tokenizer(term)
	err := db.Model(&messages).
		Where("text_indexed @> ? ", pg.Array(terms)).
		OrderExpr("? / array_length(text_indexed,1) DESC", len(terms)).
		Select()
	if err != nil {
		return nil, err
	}
	return messages, nil
}

type MessageSearch struct {
	ID          int
	Text        string
	TextIndexed []string `pg:",type:text[]"`
}

func (b *MessageSearch) BeforeInsert(ctx context.Context) (context.Context, error) {
	b.TextIndexed = tokenizer(b.Text)
	return ctx, nil
}

func (b *MessageSearch) BeforeUpdate(ctx context.Context) (context.Context, error) {
	b.TextIndexed = tokenizer(b.Text)
	return ctx, nil
}

func tokenizer(s string) []string {
	analyser := AcquireAnalyser()
	defer ReleaseAnalyser(analyser)
	tokens := analyser.Analyse(s)
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = xbytes.UnsafeString(token.Term)
	}
	return result
}
