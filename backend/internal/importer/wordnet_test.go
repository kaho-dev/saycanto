package importer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

func writeWordNetZip(t *testing.T, files map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "wn.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	_ = zw.Close()
	_ = f.Close()
	return p
}

func testWordNet(t *testing.T) *WordNet {
	t.Helper()
	wn, err := LoadWordNet(writeWordNetZip(t, map[string]string{
		"oewn/index.noun": "  1 licence header line\n" +
			"water n 6 8 ~ @ 6 0 1 2 3 4 5 6\nlight n 15 7 + @ 15 0 1\nschool n 7 6 ~ 7 0 1\n" +
			"ice_cream n 1 1 @ 1 0 1\ndrink n 5 3 @ 5 0 1\nmouse n 4 2 @ 4 0 1\nbattery n 6 2 @ 6 0 1\nverb n 2 1 @ 2 0 1\nbad n x\n",
		"oewn/index.adj":  "beautiful a 2 4 ! 2 0 1\nlight a 25 6 ! 25 0 1\n",
		"oewn/index.adv":  "deliberately r 2 1 ! 2 0 1\nlight r 1 0 1 0 1\n",
		"oewn/index.verb": "water v 4 3 ~ 4 0 1\n", // ignored: verbs come from "to ..."
		"oewn/noun.exc":   "mice mouse\n",
		"oewn/data.noun":  "ignored\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	return wn
}

func TestWordNetTypeOf(t *testing.T) {
	wn := testWordNet(t)
	cases := map[string]string{
		"water":              "noun",      // 6 noun senses; verb senses ignored for a bare meaning
		"Beautiful":          "adjective", // case-insensitive
		"light":              "",          // 15 noun / 25 adjective / 1 adverb: too split
		"ice cream":          "noun",      // multi-word lemma
		"a school":           "noun",      // leading article
		"drinks":             "noun",      // regular plural
		"batteries":          "noun",      // -ies plural
		"mice":               "noun",      // irregular plural
		"deliberately":       "adverb",
		"unknownword":        "",
		"":                   "",
		"one two three four": "",
	}
	for in, want := range cases {
		if got := wn.TypeOf(in); got != want {
			t.Errorf("TypeOf(%q) = %q, want %q", in, got, want)
		}
	}
	if singular("glass") != "glass" || singular("boxes") != "box" || singular("is") != "is" {
		t.Error("singular rules")
	}
}

func TestInferWordTypesFromWordNet(t *testing.T) {
	wn := testWordNet(t)
	entries := []dictionary.Entry{
		{Traditional: "水", English: []string{"water"}},
		{Traditional: "已有", English: []string{"water"}, WordTypes: []string{"verb"}}, // already typed: untouched
		{Traditional: "澆", English: []string{"to water", "to irrigate"}},             // verb meanings only: nothing from WordNet
		{Traditional: "美", English: []string{"beautiful", "(n.) beauty"}},            // first meanings
		{Traditional: "遠", English: []string{"far", "distant", "remote", "water"}},   // 4th meaning not looked at
		{Traditional: "光", English: []string{"light"}},                               // too split
		{Traditional: "雪糕", English: []string{"ice cream; ice-cream bar (Cantonese)"}},
	}
	if n := InferWordTypesFromWordNet(entries, wn); n != 3 {
		t.Errorf("typed %d, want 3", n)
	}
	want := [][]string{{"noun"}, {"verb"}, nil, {"adjective"}, nil, nil, {"noun"}}
	for i, e := range entries {
		if !reflect.DeepEqual(e.WordTypes, want[i]) {
			t.Errorf("%s: got %v, want %v", e.Traditional, e.WordTypes, want[i])
		}
	}
	if InferWordTypesFromWordNet(entries, nil) != 0 {
		t.Error("nil WordNet should type nothing")
	}
}

func TestLoadWordNetErrors(t *testing.T) {
	if _, err := LoadWordNet(filepath.Join(t.TempDir(), "missing.zip")); err == nil {
		t.Error("expected error for missing file")
	}
	if _, err := LoadWordNet(writeWordNetZip(t, map[string]string{"readme.txt": "x"})); err == nil {
		t.Error("expected error for zip without index files")
	}
}
