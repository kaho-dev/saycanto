package dictionary

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	seed := []Entry{
		{
			ID:           "m4-goi1",
			English:      []string{"thank you", "please", "excuse me"},
			Jyutping:     "m4 goi1",
			Traditional:  "唔該",
			Simplified:   "唔该",
			PartOfSpeech: "expression",
			Register:     "spoken",
			Examples:     []Example{{Jyutping: "m4 goi1", Cantonese: "唔該", English: "Please"}},
		},
		{
			ID:           "seoi2",
			English:      []string{"water"},
			Jyutping:     "seoi2",
			Traditional:  "水",
			Simplified:   "水",
			PartOfSpeech: "noun",
			Register:     "both",
			Classifier:   "bui1 (杯)",
		},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert entries: %v", err)
	}
	return db
}

func TestSQLiteRepositorySearchAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Search English phrase
	entries, err := repo.Search(ctx, SearchQuery{Query: "thank", Limit: 10})
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "m4-goi1" {
		t.Fatalf("expected m4-goi1, got %+v", entries)
	}

	// Search Jyutping prefix with register filter
	entries, err = repo.Search(ctx, SearchQuery{Query: "seoi", Register: "spoken", Limit: 10})
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "seoi2" {
		t.Fatalf("expected seoi2 (register 'both' matches 'spoken'), got %+v", entries)
	}

	// Sanitized FTS characters should not fail
	entries, err = repo.Search(ctx, SearchQuery{Query: `"thank"* (you):`, Limit: 10})
	if err != nil {
		t.Fatalf("sanitized search error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// GetByID found
	got, err := repo.GetByID(ctx, "m4-goi1")
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got.Jyutping != "m4 goi1" || len(got.Examples) != 1 {
		t.Fatalf("unexpected entry: %+v", got)
	}

	// GetByID not found
	_, err = repo.GetByID(ctx, "non-existent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteRepositoryList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// List all with deterministic order (spoken first, then both)
	entries, total, err := repo.List(ctx, ListQuery{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if total != 2 || len(entries) != 2 {
		t.Fatalf("expected total=2 and 2 entries, got total=%d len=%d", total, len(entries))
	}
	if entries[0].ID != "m4-goi1" || entries[1].ID != "seoi2" {
		t.Fatalf("unexpected order: %+v", entries)
	}

	// Pagination with limit & offset
	entries, total, err = repo.List(ctx, ListQuery{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("list pagination error: %v", err)
	}
	if total != 2 || len(entries) != 1 || entries[0].ID != "seoi2" {
		t.Fatalf("expected total=2 and seoi2, got total=%d entries=%+v", total, entries)
	}

	// Register filter
	entries, total, err = repo.List(ctx, ListQuery{Register: "spoken", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list register error: %v", err)
	}
	if total != 2 || len(entries) != 2 {
		t.Fatalf("expected spoken+both to match 2 entries, got total=%d len=%d", total, len(entries))
	}

	entries, total, err = repo.List(ctx, ListQuery{Register: "formal", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list formal error: %v", err)
	}
	if total != 1 || len(entries) != 1 || entries[0].ID != "seoi2" {
		t.Fatalf("expected formal+both to match seoi2, got total=%d entries=%+v", total, entries)
	}
}

func TestSQLiteRepositorySearchRanking(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	seed := []Entry{
		{ID: "zai3-seoi2", English: []string{"water restriction", "water rationing"}, Jyutping: "zai3 seoi2", Traditional: "制水", Simplified: "制水", Register: "spoken"},
		{ID: "zeng2-seoi2", English: []string{"water from a water well"}, Jyutping: "zeng2 seoi2", Traditional: "井水", Simplified: "井水", Register: "spoken"},
		{ID: "seoi2", English: []string{"water"}, Jyutping: "seoi2", Traditional: "水", Simplified: "水", Register: "both"},
		{ID: "sik6-maan5-faan6", English: []string{"to have supper dinner"}, Jyutping: "sik6 maan5 faan6", Traditional: "食晚飯", Simplified: "食晚饭", Register: "spoken"},
		{ID: "sik6-je5", English: []string{"to eat"}, Jyutping: "sik6 je5", Traditional: "食嘢", Simplified: "食嘢", Register: "spoken"},
		{ID: "gong2-sik6", English: []string{"formal eating"}, Jyutping: "gong2 sik6", Traditional: "講食", Simplified: "讲食", Register: "formal"},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	repo := NewSQLiteRepository(db)

	cases := []struct{ query, wantFirst string }{
		{"water", "seoi2"},       // exact gloss beats spoken partial matches
		{"  Water ", "seoi2"},    // case and whitespace insensitive
		{"eat", "sik6-je5"},      // "to eat" counts as exact for verbs
		{"seoi2", "seoi2"},       // exact Jyutping
		{"水", "seoi2"},           // exact headword
		{"sik6 je5", "sik6-je5"}, // exact multi-syllable Jyutping
	}
	for _, c := range cases {
		got, err := repo.Search(context.Background(), SearchQuery{Query: c.query, Limit: 10})
		if err != nil {
			t.Fatalf("%q: %v", c.query, err)
		}
		if len(got) == 0 || got[0].ID != c.wantFirst {
			ids := make([]string, len(got))
			for i, e := range got {
				ids[i] = e.ID
			}
			t.Errorf("%q: want %s first, got %v", c.query, c.wantFirst, ids)
		}
	}
}

func TestSQLiteRepositorySearchCuratedBoost(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		{ID: "hek3", English: []string{"to eat", "to consume"}, Jyutping: "hek3", Traditional: "吃", Simplified: "吃", Register: "spoken", Source: SourceCCCanto},
		{ID: "sik6-faan6", English: []string{"eat", "have a meal", "dine"}, Jyutping: "sik6 faan6", Traditional: "食飯", Simplified: "食饭", Register: "spoken", Source: SourceCustom},
		{ID: "sik6-maan5-faan6", English: []string{"to have supper dinner"}, Jyutping: "sik6 maan5 faan6", Traditional: "食晚飯", Simplified: "食晚饭", Register: "spoken", Source: SourceCCCanto},
		{ID: "sik6-je5", English: []string{"to eat"}, Jyutping: "sik6 je5", Traditional: "食嘢", Simplified: "食嘢", Register: "spoken", Source: SourceCCCanto},
		{ID: "zou2-san4", English: []string{"good morning"}, Jyutping: "zou2 san4", Traditional: "早晨", Simplified: "早晨", Register: "spoken", Source: SourceCustom},
		{ID: "faan6", English: []string{"rice", "meal"}, Jyutping: "faan6", Traditional: "飯", Simplified: "饭", Register: "spoken", Source: SourceCCCanto},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	repo := NewSQLiteRepository(db)

	ids := func(q string) []string {
		got, err := repo.Search(context.Background(), SearchQuery{Query: q, Limit: 10})
		if err != nil {
			t.Fatalf("%q: %v", q, err)
		}
		out := make([]string, len(got))
		for i, e := range got {
			out[i] = e.ID
		}
		return out
	}

	// Among exact matches, the curated entry comes first.
	if got := ids("eat"); got[0] != "sik6-faan6" {
		t.Errorf("eat: want curated sik6-faan6 first, got %v", got)
	}
	// An exact CC-Canto match still beats a curated partial match ("faan6" is in sik6-faan6's Jyutping).
	if got := ids("faan6"); got[0] != "faan6" || got[1] != "sik6-faan6" {
		t.Errorf("faan6: want [faan6 sik6-faan6 ...], got %v", got)
	}
	// Source survives the round trip.
	e, err := repo.GetByID(context.Background(), "zou2-san4")
	if err != nil || e.Source != SourceCustom {
		t.Errorf("GetByID source: got %+v, err %v", e, err)
	}
}

// A database built before the source column existed is upgraded in place.
func TestInitSchemaAddsSourceColumnToOldDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE entries (id TEXT PRIMARY KEY, english_json TEXT NOT NULL, jyutping TEXT NOT NULL,
		traditional TEXT NOT NULL, simplified TEXT NOT NULL, part_of_speech TEXT NOT NULL DEFAULT '',
		register TEXT NOT NULL DEFAULT '', classifier TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '',
		examples_json TEXT NOT NULL DEFAULT '[]');
		INSERT INTO entries (id, english_json, jyutping, traditional, simplified) VALUES ('seoi2', '["water"]', 'seoi2', '水', '水');`); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	for i := 0; i < 2; i++ { // idempotent
		if err := InitSchema(db); err != nil {
			t.Fatalf("InitSchema run %d: %v", i+1, err)
		}
	}
	e, err := NewSQLiteRepository(db).GetByID(context.Background(), "seoi2")
	if err != nil || e.Source != "" {
		t.Fatalf("old row after upgrade: %+v, err %v", e, err)
	}
}

func setupEmptyDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return db
}

// Starter words (the unfiltered list) show curated entries before bulk CC-Canto ones.
func TestSQLiteRepositoryListCuratedFirst(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		{ID: "1-mok1", English: []string{"filefish"}, Jyutping: "1 mok1", Traditional: "1剝", Simplified: "1剥", Register: "spoken", Source: SourceCCCanto},
		{ID: "seoi2", English: []string{"water"}, Jyutping: "seoi2", Traditional: "水", Simplified: "水", Register: "both", Source: SourceCustom},
		{ID: "aa1", English: []string{"oh"}, Jyutping: "aa1", Traditional: "吖", Simplified: "吖", Register: "spoken", Source: SourceCCCanto},
		{ID: "m4-goi1", English: []string{"thank you"}, Jyutping: "m4 goi1", Traditional: "唔該", Simplified: "唔该", Register: "spoken", Source: SourceCustom},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, total, err := NewSQLiteRepository(db).List(context.Background(), ListQuery{Limit: 10})
	if err != nil || total != 4 {
		t.Fatalf("list: total %d, err %v", total, err)
	}
	want := []string{"m4-goi1", "seoi2", "1-mok1", "aa1"}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("position %d: want %s, got %s", i, id, got[i].ID)
		}
	}
}

func TestSQLiteRepositoryWordTypes(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		{ID: "tung4", English: []string{"and", "with"}, Jyutping: "tung4", Traditional: "同", Simplified: "同", Register: "spoken", WordTypes: []string{"preposition", "conjunction"}},
		{ID: "tung4-maai4", English: []string{"and"}, Jyutping: "tung4 maai4", Traditional: "同埋", Simplified: "同埋", Register: "spoken", WordTypes: []string{"conjunction"}},
		{ID: "kap6", English: []string{"and", "to reach"}, Jyutping: "kap6", Traditional: "及", Simplified: "及", Register: "formal", WordTypes: []string{"verb", "conjunction"}},
		{ID: "wo4", English: []string{"and", "harmony"}, Jyutping: "wo4", Traditional: "和", Simplified: "和", Register: "formal"},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatal(err)
	}
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	got, err := repo.Search(ctx, SearchQuery{Query: "and", WordType: "conjunction", Limit: 10})
	if err != nil || len(got) != 3 {
		t.Fatalf("conjunction filter: %d results, err %v", len(got), err)
	}
	e, _ := repo.GetByID(ctx, "tung4")
	if !reflect.DeepEqual(e.WordTypes, []string{"preposition", "conjunction"}) {
		t.Errorf("round trip word types = %v", e.WordTypes)
	}
	if w, _ := repo.GetByID(ctx, "wo4"); w.WordTypes != nil {
		t.Errorf("untyped entry = %v", w.WordTypes)
	}

	counts, err := repo.CountTypes(ctx, SearchQuery{Query: "and", WordType: "verb"})
	if err != nil {
		t.Fatal(err)
	}
	want := TypeCounts{Total: 4, Untyped: 1, ByType: map[string]int{"conjunction": 3, "preposition": 1, "verb": 1}}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("counts = %+v, want %+v", counts, want)
	}
	if c, _ := repo.CountTypes(ctx, SearchQuery{Query: "and", Register: "formal"}); c.Total != 2 {
		t.Errorf("formal total = %d", c.Total)
	}
	if c, _ := repo.CountTypes(ctx, SearchQuery{Query: "  "}); c.Total != 0 {
		t.Errorf("empty query total = %d", c.Total)
	}

	list, total, err := repo.List(ctx, ListQuery{WordType: "conjunction", Register: "spoken", Limit: 10})
	if err != nil || total != 2 || len(list) != 2 {
		t.Errorf("list: total %d len %d err %v", total, len(list), err)
	}
}

func TestCountTypesIgnoresPrefixMatches(t *testing.T) {
	db := setupEmptyDB(t)
	seed := []Entry{
		{ID: "so2-ji5", English: []string{"so", "therefore"}, Jyutping: "so2 ji5", Traditional: "所以", Simplified: "所以", WordTypes: []string{"conjunction"}},
		{ID: "di1", English: []string{"some"}, Jyutping: "di1", Traditional: "啲", Simplified: "啲", WordTypes: []string{"classifier"}},
		{ID: "je5", English: []string{"thing"}, Jyutping: "je5", Traditional: "嘢", Simplified: "嘢"},
		{ID: "je5-jam2", English: []string{"drinks"}, Jyutping: "je5 jam2", Traditional: "嘢飲", Simplified: "嘢饮"},
	}
	if err := InsertEntries(context.Background(), db, seed); err != nil {
		t.Fatal(err)
	}
	repo := NewSQLiteRepository(db)
	c, err := repo.CountTypes(context.Background(), SearchQuery{Query: "so"})
	if err != nil || c.Total != 1 || c.ByType["conjunction"] != 1 || c.ByType["classifier"] != 0 {
		t.Errorf("so: %+v err %v", c, err)
	}
	// Chinese characters still match as a prefix: 嘢 counts 嘢 and 嘢飲.
	if c, _ := repo.CountTypes(context.Background(), SearchQuery{Query: "嘢"}); c.Total != 2 || c.Untyped != 2 {
		t.Errorf("嘢: %+v", c)
	}
	// With a type filter the list uses whole words too, so it agrees with the count.
	if got, _ := repo.Search(context.Background(), SearchQuery{Query: "so", WordType: "classifier", Limit: 10}); len(got) != 0 {
		t.Errorf("so + classifier should not list 啲 (some): %+v", got)
	}
	if got, _ := repo.Search(context.Background(), SearchQuery{Query: "so", Limit: 10}); len(got) != 2 {
		t.Errorf("unfiltered so should still prefix-match some: %d", len(got))
	}
	if got := buildFTSWholeWordQuery(`m4 "goi1" 唔`); got != `"m4" "goi1" "唔"*` {
		t.Errorf("whole-word FTS = %s", got)
	}
}
