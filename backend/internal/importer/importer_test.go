package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
	_ "modernc.org/sqlite"
)

func TestParseCCCanto(t *testing.T) {
	input := `
# CC-Canto sample
唔該 唔該 [wu1 gai1] {m4 goi1} /please/thank you/
唔該 唔該 [wu1 gai1] {m4 goi1} /excuse me/
malformed line without brackets
`
	entries, err := ParseCCCanto(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].ID != "m4-goi1" || entries[1].ID != "m4-goi1-2" {
		t.Fatalf("unexpected deduplicated IDs: %s, %s", entries[0].ID, entries[1].ID)
	}
	if entries[0].Jyutping != "m4 goi1" || len(entries[0].English) != 2 {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
}

func TestParseCustomJSONAndMergeAndBuild(t *testing.T) {
	customJSON := `[
		{
			"id": "m4-goi1",
			"english": ["thank you", "please"],
			"jyutping": "m4 goi1",
			"traditional": "唔該",
			"simplified": "唔该",
			"register": "spoken",
			"notes": "Curated note"
		},
		{
			"id": "seoi2",
			"english": ["water"],
			"jyutping": "seoi2",
			"traditional": "水",
			"simplified": "水",
			"register": "both"
		}
	]`

	overlay, err := ParseCustomJSON(strings.NewReader(customJSON))
	if err != nil {
		t.Fatalf("ParseCustomJSON error: %v", err)
	}

	base := []dictionary.Entry{
		{
			ID:          "m4-goi1",
			English:     []string{"raw"},
			Jyutping:    "m4 goi1",
			Traditional: "唔該",
			Simplified:  "唔该",
		},
	}

	merged := MergeEntries(base, overlay)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged entries, got %d", len(merged))
	}
	if merged[0].Notes != "Curated note" {
		t.Fatalf("expected overlay to override base entry, got %+v", merged[0])
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := BuildDatabase(context.Background(), db, merged); err != nil {
		t.Fatalf("BuildDatabase error: %v", err)
	}

	repo := dictionary.NewSQLiteRepository(db)
	got, err := repo.GetByID(context.Background(), "seoi2")
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got.Jyutping != "seoi2" {
		t.Fatalf("expected seoi2, got %s", got.Jyutping)
	}
}

// Regression: duplicate CC-Canto readings plus a curated entry used to produce two
// rows with the same id and fail the build with "UNIQUE constraint failed: entries.id".
func TestMergeEntriesCCCantoDuplicatesWithOverlayBuilds(t *testing.T) {
	base, err := ParseCCCanto(strings.NewReader(
		"唔該 唔該 [wu1 gai1] {m4 goi1} /please/\n" +
			"唔該 唔該 [wu1 gai1] {m4 goi1} /excuse me/\n" +
			"嘸該 嘸该 [wu2 gai1] {m4 goi1} /variant/\n"))
	if err != nil {
		t.Fatalf("ParseCCCanto: %v", err)
	}
	overlay := []dictionary.Entry{{
		ID: "m4-goi1", English: []string{"thank you"}, Jyutping: "m4 goi1",
		Traditional: "唔該", Simplified: "唔该", Notes: "curated",
	}}

	merged := MergeEntries(base, overlay)
	if len(merged) != 2 {
		t.Fatalf("expected curated entry to replace both 唔該 lines (2 entries), got %d: %+v", len(merged), merged)
	}
	if merged[0].ID != "m4-goi1" || merged[0].Notes != "curated" {
		t.Fatalf("expected curated m4-goi1 first, got %+v", merged[0])
	}
	if merged[1].Traditional != "嘸該" || merged[1].ID == "m4-goi1" {
		t.Fatalf("expected variant with a distinct id, got %+v", merged[1])
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if err := BuildDatabase(context.Background(), db, merged); err != nil {
		t.Fatalf("BuildDatabase: %v", err)
	}
}

func TestMergeEntriesRenumbersBaseIDTakenByOverlay(t *testing.T) {
	base := []dictionary.Entry{
		{ID: "hou2", Jyutping: "hou2", Traditional: "好"},
		{ID: "hou2-2", Jyutping: "hou2", Traditional: "號"},
	}
	// Curated entry for a different word that claims the base ids.
	overlay := []dictionary.Entry{
		{ID: "hou2", Jyutping: "hou2", Traditional: "毫"},
		{ID: "hou2-2", Jyutping: "hou2", Traditional: "蠔"},
	}

	merged := MergeEntries(base, overlay)
	got := make([]string, len(merged))
	seen := map[string]bool{}
	for i, e := range merged {
		got[i] = e.Traditional + "=" + e.ID
		if seen[e.ID] {
			t.Fatalf("duplicate id %q in %v", e.ID, got)
		}
		seen[e.ID] = true
	}
	want := []string{"好=hou2-3", "號=hou2-4", "毫=hou2", "蠔=hou2-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestValidateOverlay(t *testing.T) {
	ok := []dictionary.Entry{{ID: "a"}, {ID: "b"}}
	if err := ValidateOverlay(ok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateOverlay([]dictionary.Entry{{ID: "a"}, {ID: "a"}}); err == nil {
		t.Fatal("expected duplicate id error")
	}
	if err := ValidateOverlay([]dictionary.Entry{{ID: " ", Traditional: "水"}}); err == nil {
		t.Fatal("expected missing id error")
	}
}

func TestStripNumericSuffix(t *testing.T) {
	cases := map[string]string{"m4-goi1-2": "m4-goi1", "m4-goi1": "m4-goi1", "seoi2": "seoi2", "-3": "-3"}
	for in, want := range cases {
		if got := stripNumericSuffix(in); got != want {
			t.Errorf("stripNumericSuffix(%q) = %q, want %q", in, got, want)
		}
	}
}

// Regression: lines with a trailing source note were silently dropped.
func TestParseCCCantoTrailingComment(t *testing.T) {
	input := "丁 丁 [ding1] {ding1} /fourth of the ten Heavenly Stems/a surname/ # adapted from cc-cedict\n"
	entries, err := ParseCCCanto(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if got := entries[0].English; len(got) != 2 || got[1] != "a surname" {
		t.Fatalf("unexpected definitions: %q", got)
	}
}

// Regression: Cantonese-only words have an empty Mandarin reading "[]".
func TestParseCCCantoEmptyPinyin(t *testing.T) {
	entries, err := ParseCCCanto(strings.NewReader("𠌥低 𠆿低 [] {wu1 dai1} /(verb) to stoop; to bend down/\n"))
	if err != nil || len(entries) != 1 || entries[0].Jyutping != "wu1 dai1" {
		t.Fatalf("got %+v, err %v", entries, err)
	}
}

func TestParsersSetSource(t *testing.T) {
	cc, _ := ParseCCCanto(strings.NewReader("水 水 [shui3] {seoi2} /water/\n"))
	custom, _ := ParseCustomJSON(strings.NewReader(`[{"id":"seoi2","english":["water"],"jyutping":"seoi2","traditional":"水","simplified":"水","source":"cc-canto"}]`))
	if cc[0].Source != dictionary.SourceCCCanto || custom[0].Source != dictionary.SourceCustom {
		t.Fatalf("sources: cc=%q custom=%q", cc[0].Source, custom[0].Source)
	}
}
