package importer

import (
	"reflect"
	"testing"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

func TestInferWordTypes(t *testing.T) {
	hk := CantoneseFreq{Tags: map[string][]string{
		"佢|keoi5":      {"r"},
		"但係|daan6hai6": {"c"},
		"嘅|ge3":        {"u", "y"},
		"去|heoi3":      {"v"},
		"阿嬸|aa3sam2":   {"n"},
		"嗰陣|go2zan6":   {"xn"}, // unmapped tag: no type
	}}
	entries := []dictionary.Entry{
		{Traditional: "同", Jyutping: "tung4", PartOfSpeech: "Conjunction", English: []string{"and", "with"}},
		{Traditional: "佢", Jyutping: "keoi5", English: []string{"he; she; it"}},
		{Traditional: "但係", Jyutping: "daan6 hai6", English: []string{"(conjunction) but"}},
		{Traditional: "嘅", Jyutping: "ge3", English: []string{"possessive particle"}},
		{Traditional: "去", Jyutping: "heoi3", English: []string{"to go", "to go to (a place)"}},
		{Traditional: "條", Jyutping: "tiu4", English: []string{"strip", "classifier for long thin things"}},
		{Traditional: "唔", Jyutping: "m4", English: []string{"(phrase; adverb, verb) don't"}},
		{Traditional: "啦", Jyutping: "laa1", English: []string{"sentence-final particle, indicating assertion"}},
		{Traditional: "畫蛇添足", Jyutping: "waak6 se4 tim1 zuk1", English: []string{"lit. draw legs on a snake (idiom); fig. to ruin the effect", "to overdo it"}},
		{Traditional: "滾水", Jyutping: "gwan2 seoi2", English: []string{"(n.) boiling water"}},
		{Traditional: "阿嬸", Jyutping: "aa3 sam2", English: []string{"auntie"}},
		{Traditional: "嗰陣", Jyutping: "go2 zan6", English: []string{"at that time"}},
		{Traditional: "怪", Jyutping: "gwaai3", English: []string{"(xyz) strange"}}, // unknown label
		{Traditional: "打簿", Jyutping: "daa2 bou2", English: []string{"verb; update bank passbook"}},
	}
	if n := InferWordTypes(entries, hk); n != 12 {
		t.Errorf("typed %d entries, want 12", n)
	}
	want := map[string][]string{
		"同":    {"conjunction"},
		"佢":    {"pronoun"},
		"但係":   {"conjunction"},
		"嘅":    {"particle"},
		"去":    {"verb"},
		"條":    {"noun"}, // only via label? no: "strip" is bare; classifier from pattern
		"唔":    {"verb", "adverb", "expression"},
		"啦":    {"particle"},
		"畫蛇添足": {"expression"},
		"滾水":   {"noun"},
		"阿嬸":   {"noun"},
		"嗰陣":   nil,
		"怪":    nil,
		"打簿":   {"verb"},
	}
	want["條"] = []string{"classifier"}
	for _, e := range entries {
		if !reflect.DeepEqual(e.WordTypes, want[e.Traditional]) {
			t.Errorf("%s: got %v, want %v", e.Traditional, e.WordTypes, want[e.Traditional])
		}
	}
}

func TestLoadHKCanCorTags(t *testing.T) {
	transcript := "但係/c/daan6hai6/\n但係/c/daan6hai6/\n但係/c/daan6hai6/\n但係/c/daan6hai6/\n但係/d/daan6hai6/\n" +
		"去/v/heoi3/\n去/v/heoi3/\n去/v/heoi3/\n去/v/heoi3/\n去/v/heoi3/\n去/p/heoi3/\n"
	hk, err := LoadHKCanCor(writeHKCanCorZip(t, map[string]string{"utf8/a": transcript}))
	if err != nil {
		t.Fatal(err)
	}
	// 但係: c 4/5 and d 1/5 (20%) both kept; 去: p 1/6 (17%) dropped.
	if got := hk.Tags["但係|daan6hai6"]; !reflect.DeepEqual(got, []string{"c", "d"}) {
		t.Errorf("但係 tags = %v", got)
	}
	if got := hk.Tags["去|heoi3"]; !reflect.DeepEqual(got, []string{"v"}) {
		t.Errorf("去 tags = %v", got)
	}
}
