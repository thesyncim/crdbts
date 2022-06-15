package crdbts

import (
	"testing"

	"github.com/GetStream/kit/v3/pgutil"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type TestCase struct {
	Records    []*MessageSearch
	SearchTerm string
	Expected   []int
}

func (tc TestCase) Run(t testing.TB, db *pg.DB) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	err = orm.NewQuery(tx).Model(&MessageSearch{}).CreateTable(&orm.CreateTableOptions{})
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("create index concurrently text_indexed on message_searches using gin(text_indexed)")
	if err != nil {
		panic(err)
	}
	_, err = tx.Model(&tc.Records).Insert()
	if err != nil {
		t.Fatal(err)
	}

	searchResults, err := Search(tx, tc.SearchTerm)
	if err != nil {
		t.Fatal(err)
	}
	if len(searchResults) != len(tc.Expected) {
		t.Fatalf("Expected %d results, got %d", len(tc.Expected), len(searchResults))
	}
	for i := range searchResults {
		if searchResults[i].ID != tc.Expected[i] {
			t.Errorf("Expected %d, got %d", tc.Expected[i], searchResults[i].ID)
		}
	}
}

func TestTokenizer(t *testing.T) {
	var tests = []struct {
		text     string
		expected []string
	}{
		{
			text: "the quick brown fox jumps over the lazy dog",
			expected: []string{
				"quick", "brown", "fox", "jump", "lazi", "dog",
			},
		},
		{
			text: "🚲🚲🚲🚲",
			expected: []string{
				"🚲",
			},
		},
		{
			text: "marcelo's 🚲",
			expected: []string{
				"marcelo", "🚲",
			},
		},
		{
			text: "thesyncim@mail.abc",
			expected: []string{
				"thesyncim", "mail.abc",
			},
		},
		{
			text: "#thesyncim",
			expected: []string{
				"thesyncim",
			},
		},
		{
			text: "http://google.pt/search?q=something",
			expected: []string{
				"http", "google.pt", "search", "q", "=", "someth",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			tokens := tokenizer(test.text)
			if len(tokens) != len(test.expected) {
				t.Logf("%#+v", tokens)
				t.Fatalf("Expected %d tokens, got %d", len(test.expected), len(tokens))
			}
			for i := range tokens {
				if tokens[i] != test.expected[i] {
					t.Errorf("Expected %s, got %s", test.expected[i], tokens[i])
				}
			}
		})
	}

}

func TestSearch(t *testing.T) {
	db, err := pgutil.NewORM("host=localhost user=chat dbname=chat sslmode=disable port=26257", pgutil.PGSettings{})
	if err != nil {
		panic(err)
	}
	var testCases = []TestCase{
		{
			Records: []*MessageSearch{
				{ID: 1, Text: "the quick brown fox jumps over the lazy dog"},
				{ID: 2, Text: "lazy fox"},
			},
			SearchTerm: "fox lazy",
			Expected:   []int{2, 1},
		},
		{
			Records: []*MessageSearch{
				{ID: 1, Text: "love to ride my 🚲"},
				{ID: 2, Text: "love to ride my"},
			},
			SearchTerm: "my 🚲",
			Expected:   []int{1},
		},
		{
			Records: []*MessageSearch{
				{ID: 1, Text: "github.com/thesyncim"},
				{ID: 2, Text: "thesyncim@mail.abc"},
				{ID: 3, Text: ""},
			},
			SearchTerm: "thesyncim",
			Expected:   []int{1, 2},
		},
		{
			Records:    []*MessageSearch{{ID: 100, Text: "我要一杯啤酒"}},
			SearchTerm: "我要",
			Expected:   []int{100},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].SearchTerm, func(t *testing.T) {
			testCases[i].Run(t, db)
		})
	}
}

func BenchmarkStemParallel(b *testing.B) {
	b.ReportAllocs()
	var text = "the quick brown fox jumps over the lazy dog"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := tokenizer(text)
			if len(v) != 6 {
				b.Errorf("Expected length of 6, got %d", len(v))
			}
			if v[0] != "quick" || v[1] != "brown" || v[2] != "fox" || v[3] != "jump" || v[4] != "lazi" || v[5] != "dog" {
				b.Errorf("Expected %v, got %v", []string{"quick", "brown", "fox", "jump", "lazi", "dog"}, v)
			}
			if text != "the quick brown fox jumps over the lazy dog" {
				b.Errorf("Expected %s, got %s", "the quick brown fox jumps over the lazy dog", text)
			}
		}
	})
}
