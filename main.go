package crdbts

import (
	"context"

	"github.com/GetStream/kit/v3/xbytes"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func Search(db orm.DB, term string) ([]*MessageSearch, error) {
	var messages []*MessageSearch
	err := db.Model(&messages).Where("text_indexed @> ? ", pg.Array(stem(term))).Select()
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
	b.TextIndexed = stem(b.Text)
	return ctx, nil
}

func (b *MessageSearch) BeforeUpdate(ctx context.Context) (context.Context, error) {
	b.TextIndexed = stem(b.Text)
	return ctx, nil
}

func stem(s string) []string {
	if len(s) == 0 {
		return nil
	}

	analiser := AcquireAnalyser()
	defer ReleaseAnalyser(analiser)
	tokens := analiser.Analyse(s)
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = xbytes.UnsafeString(token.Term)
	}
	return result
}
