package dictionary

import (
	"context"
	"reflect"
	"testing"
)

func TestGlossKeys(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"(conjunction) however; but;"}, "|however|but|"},
		{[]string{"to eat", "to consume"}, "|to eat|eat|to consume|consume|"},
		{[]string{"regular; usual; usually; normally"}, "|regular|usual|usually|normally|"},
		{[]string{"thank you very much; thank you so much [colloquial]"}, "|thank you very much|thank you so much|"},
		{[]string{"Excuse me!", "  Thank   You  "}, "|excuse me|thank you|"},
		{[]string{"(of water) deep and clear"}, "|deep and clear|"},
		{[]string{"one or two, a little"}, "|one or two|a little|"},
		{[]string{"(noun)"}, ""},
		{[]string{"(adjective) 1. still; 2. calm (adverb) 1. or;", "1) to fix"}, "|still|calm|or|to fix|fix|"},
		{[]string{"version 2.0", "top 10. list"}, "|version 2.0|top|list|"},
		{nil, ""},
	}
	for _, c := range cases {
		if got := glossKeys(c.in); got != c.want {
			t.Errorf("glossKeys(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestGlossWordsAndQueryWords(t *testing.T) {
	got := glossWords([]string{"(adjective) Surprise (usually used in response)", "it's fine; roast-meat"})
	if want := " surprise it's fine roast meat "; got != want {
		t.Errorf("glossWords = %q, want %q", got, want)
	}
	if glossWords([]string{"[colloquial]"}) != "" {
		t.Error("expected empty gloss words")
	}
	if got := queryWords("  But, USUALLY? m4 goi1 "); !reflect.DeepEqual(got, []string{"but", "usually", "m4", "goi1"}) {
		t.Errorf("queryWords = %q", got)
	}
	if got := exactKey(" Usually. "); got != "usually" {
		t.Errorf("exactKey = %q", got)
	}
}

func TestSearchRankingTiers(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		{ID: "lo1-pet1", English: []string{"butt; buttocks [colloquial]"}, Jyutping: "lo1 pet1", Traditional: "籮Ｐｅｔ", Simplified: "籮Ｐｅｔ", Source: SourceCCCanto},
		{ID: "tou4", English: []string{"to slaughter", "butcher"}, Jyutping: "tou4", Traditional: "屠", Simplified: "屠", Source: SourceCCCanto},
		{ID: "zau1-si4", English: []string{"often or usually (negatively)"}, Jyutping: "zau1 si4", Traditional: "周時", Simplified: "周时", Source: SourceCCCanto},
		{ID: "hei1-kei4", English: []string{"(adjective) Surprise (usually used in response)"}, Jyutping: "hei1 kei4", Traditional: "稀奇", Simplified: "稀奇", Source: SourceCCCanto},
		{ID: "daan6-hai6", English: []string{"(conjunction) but"}, Jyutping: "daan6 hai6", Traditional: "但係", Simplified: "但系", Source: SourceCCCanto},
		{ID: "zi1-bat1-gwo3", English: []string{"(conjunction) however; but;"}, Jyutping: "zi1 bat1 gwo3", Traditional: "之不過", Simplified: "之不过", Source: SourceCCCanto},
		{ID: "tung1-soeng4", English: []string{"regular; usual; normal; usually; normally"}, Jyutping: "tung1 soeng4", Traditional: "通常", Simplified: "通常", Source: SourceCCCanto},
		{ID: "m4-goi1", English: []string{"thank you"}, Jyutping: "m4 goi1", Traditional: "唔該", Simplified: "唔该", Source: SourceCCCanto},
		{ID: "m4-goi1-ze3-ze3", English: []string{"excuse me (politely)"}, Jyutping: "m4 goi1 ze3 ze3", Traditional: "唔該借借", Simplified: "唔该借借", Source: SourceCCCanto},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	repo := NewSQLiteRepository(db)
	ids := func(q string) []string {
		got, err := repo.Search(context.Background(), SearchQuery{Query: q, Limit: 20})
		if err != nil {
			t.Fatalf("%q: %v", q, err)
		}
		out := make([]string, len(got))
		for i, e := range got {
			out[i] = e.ID
		}
		return out
	}
	before := func(list []string, a, b string) bool {
		ia, ib := -1, -1
		for i, id := range list {
			if id == a {
				ia = i
			}
			if id == b {
				ib = i
			}
		}
		return ia >= 0 && ib >= 0 && ia < ib
	}

	but := ids("but")
	for _, partial := range []string{"lo1-pet1", "tou4"} {
		for _, exact := range []string{"daan6-hai6", "zi1-bat1-gwo3"} {
			if !before(but, exact, partial) {
				t.Errorf("but: want exact %s before prefix match %s, got %v", exact, partial, but)
			}
		}
	}

	usually := ids("usually")
	if usually[0] != "tung1-soeng4" {
		t.Errorf("usually: want 通常 first, got %v", usually)
	}
	// A whole-word mention beats one that only appears inside a bracketed aside.
	if !before(usually, "zau1-si4", "hei1-kei4") {
		t.Errorf("usually: want 周時 before 稀奇, got %v", usually)
	}

	// Whole Jyutping syllables beat prefixes; the exact reading comes first.
	if got := ids("m4 goi1"); got[0] != "m4-goi1" || got[1] != "m4-goi1-ze3-ze3" {
		t.Errorf("m4 goi1: got %v", got)
	}
	// Punctuation and case in the query do not matter.
	if got := ids("But?"); got[0] != "daan6-hai6" && got[0] != "zi1-bat1-gwo3" {
		t.Errorf("But?: got %v", got)
	}
}

func TestSearchPrefersCommonWords(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		// Rare literary word with one meaning vs. the everyday word with many.
		{ID: "sou3-jat6", English: []string{"usually"}, Jyutping: "sou3 jat6", Traditional: "素日", Simplified: "素日", Source: SourceCEDICT, Frequency: 2.9},
		{ID: "tung1-soeng4", English: []string{"regular; usual; normal", "usually; normally"}, Jyutping: "tung1 soeng4", Traditional: "通常", Simplified: "通常", Source: SourceCEDICT, Frequency: 5.4},
		// A whole-word match that is very common must not jump above an exact match.
		{ID: "siu1-juk6", English: []string{"roast meat, usually pork"}, Jyutping: "siu1 juk6", Traditional: "燒肉", Simplified: "烧肉", Source: SourceCCCanto, Frequency: 6.5},
		// Curated still beats frequency.
		{ID: "pin4-si4", English: []string{"usually"}, Jyutping: "ping4 si4", Traditional: "平時", Simplified: "平时", Source: SourceCustom, Frequency: 0},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, err := NewSQLiteRepository(db).Search(context.Background(), SearchQuery{Query: "usually", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, e := range got {
		ids = append(ids, e.ID)
	}
	want := []string{"pin4-si4", "tung1-soeng4", "sou3-jat6", "siu1-juk6"}
	if !reflect.DeepEqual(ids, want) {
		t.Errorf("got %v, want %v", ids, want)
	}
}

func TestSearchPrefersMainMeaning(t *testing.T) {
	db := setupEmptyDB(t)
	buried := []string{"important", "vital", "to want", "to ask for", "will", "going to", "may", "must", "probably", "need", "should", "essential", "key", "if"}
	seed := []Entry{
		{ID: "jiu3", English: buried, Jyutping: "jiu3", Traditional: "要", Simplified: "要", Source: SourceCCCanto, Frequency: 6.7},
		{ID: "joek6-gwo2", English: []string{"if", "in the event"}, Jyutping: "joek6 gwo2", Traditional: "若果", Simplified: "若果", Source: SourceCCCanto, Frequency: 2.7},
		{ID: "jyu4-gwo2", English: []string{"if", "in case"}, Jyutping: "jyu4 gwo2", Traditional: "如果", Simplified: "如果", Source: SourceCEDICT, Frequency: 6.0},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, err := NewSQLiteRepository(db).Search(context.Background(), SearchQuery{Query: "if", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, e := range got {
		ids = append(ids, e.ID)
	}
	// 要: 6.7 - 0.3 × 15 (keys incl. "want", "ask for"; about where "if" sits in the real entry) is far below the words whose first meaning is "if".
	if want := []string{"jyu4-gwo2", "joek6-gwo2", "jiu3"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("got %v, want %v", ids, want)
	}
}
