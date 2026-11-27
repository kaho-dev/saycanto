package importer

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// mpStr / mpArr build tiny MessagePack values for tests.
func mpStr(s string) []byte {
	if len(s) < 32 {
		return append([]byte{0xa0 | byte(len(s))}, s...)
	}
	return append([]byte{0xd9, byte(len(s))}, s...)
}

func mpArr(items ...[]byte) []byte {
	out := []byte{0xdc, 0, byte(len(items))} // array16
	for _, it := range items {
		out = append(out, it...)
	}
	return out
}

func gzipBytes(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(b)
	_ = zw.Close()
	return buf.Bytes()
}

func TestLoadWordfreq(t *testing.T) {
	header := append([]byte{0x82}, append(append(mpStr("format"), mpStr("cB")...), append(mpStr("version"), 0x01)...)...)
	// bucket 0 → Zipf 9.00, bucket 1 → 8.99, bucket 2 → 8.98
	data := mpArr(header, mpArr(mpStr("的")), mpArr(), mpArr(mpStr("通常"), mpStr("的")))
	table, err := LoadWordfreq(bytes.NewReader(gzipBytes(t, data)))
	if err != nil {
		t.Fatal(err)
	}
	if table["的"] != 9 {
		t.Errorf("的 = %v, want 9 (first bucket wins)", table["的"])
	}
	if math.Abs(table["通常"]-8.98) > 1e-9 {
		t.Errorf("通常 = %v, want 8.98", table["通常"])
	}

	bad := map[string][]byte{
		"not gzip":   []byte("plain"),
		"not a list": gzipBytes(t, mpStr("x")),
		"bad header": gzipBytes(t, mpArr(mpStr("x"))),
		"bad bucket": gzipBytes(t, mpArr(header, mpStr("x"))),
		"truncated":  gzipBytes(t, []byte{0xdc, 0, 5}),
	}
	for name, b := range bad {
		if _, err := LoadWordfreq(bytes.NewReader(b)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestDecodeMsgpackTypes(t *testing.T) {
	cases := []struct {
		in   []byte
		want any
	}{
		{[]byte{0x05}, int64(5)},
		{[]byte{0xff}, int64(-1)},
		{[]byte{0xc0}, nil},
		{[]byte{0xc2}, false},
		{[]byte{0xc3}, true},
		{[]byte{0xcc, 200}, int64(200)},
		{[]byte{0xd0, 0xfe}, int64(-2)},
		{[]byte{0xcd, 1, 0}, int64(256)},
		{[]byte{0xd1, 0xff, 0xfe}, int64(-2)},
		{[]byte{0xce, 0, 1, 0, 0}, int64(65536)},
		{[]byte{0xd2, 0xff, 0xff, 0xff, 0xfe}, int64(-2)},
		{append([]byte{0xda, 0, 2}, "ab"...), "ab"},
		{append([]byte{0xdb, 0, 0, 0, 1}, "c"...), "c"},
		{[]byte{0x91, 0x01}, []any{int64(1)}},
		{[]byte{0xdd, 0, 0, 0, 1, 0x02}, []any{int64(2)}},
		{append([]byte{0xde, 0, 1}, append(mpStr("k"), 0x03)...), map[string]any{"k": int64(3)}},
		{append([]byte{0xdf, 0, 0, 0, 1}, append(mpStr("k"), 0x04)...), map[string]any{"k": int64(4)}},
	}
	for _, c := range cases {
		got, err := decodeMsgpack(bufio.NewReader(bytes.NewReader(c.in)))
		if err != nil {
			t.Errorf("% x: %v", c.in, err)
			continue
		}
		if !equalAny(got, c.want) {
			t.Errorf("% x: got %#v, want %#v", c.in, got, c.want)
		}
	}
	for _, in := range [][]byte{{0xc1}, {0x81, 0x01, 0x01}, {}, {0xcc}, {0xd9}, {0xdc, 0}, {0xde, 0}, {0xa3, 'a'}, {0x81}, {0x81, 0xa1, 'k'}} {
		if _, err := decodeMsgpack(bufio.NewReader(bytes.NewReader(in))); err == nil {
			t.Errorf("% x: expected error", in)
		}
	}
}

func equalAny(a, b any) bool {
	switch av := a.(type) {
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !equalAny(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k := range av {
			if !equalAny(av[k], bv[k]) {
				return false
			}
		}
		return true
	}
	return a == b
}

func writeHKCanCorZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hkcancor.zip")
	f, err := os.Create(path)
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
	return path
}

func TestLoadHKCanCorAndApply(t *testing.T) {
	transcript := strings.Join([]string{
		"<sent_tag>",
		"但係/c/daan6hai6/", "但係/c/daan6hai6/", "但係/c/daan6hai6/",
		"行/v/hang4/", "行/v/hang4/", "行/q/hong4/",
		"嘢/n/je5/", "嘢/n/je5/",
		"罕見/a/hon2gin3/",
		"。/w/VQ1/", "。/w/VQ1/",
		"</sent_tag>",
	}, "\n")
	path := writeHKCanCorZip(t, map[string]string{
		"utf8/FC-001": transcript,
		"utf8/FC-002": "嘢/n/je5/\n",
		"README":      "但係/c/daan6hai6/\n", // outside utf8/: ignored
	})
	hk, err := LoadHKCanCor(path)
	if err != nil {
		t.Fatal(err)
	}
	if hk.Words != 10 {
		t.Errorf("Words = %d, want 10 (punctuation and README ignored)", hk.Words)
	}
	if _, ok := hk.ByReading["罕見|hon2gin3"]; ok {
		t.Error("words seen once should be dropped")
	}
	if _, ok := hk.ByReading["行|hong4"]; ok {
		t.Error("行 hong4 was seen once and should be dropped")
	}
	if got, want := hk.ByReading["但係|daan6hai6"], math.Log10(3.0/10*1e9); math.Abs(got-want) > 1e-9 {
		t.Errorf("但係 = %v, want %v", got, want)
	}

	wf := WordfreqTable{"但是": 6.1, "行": 5.5, "通常": 5.4}
	entries := []dictionary.Entry{
		{Traditional: "但係", Simplified: "但系", Jyutping: "daan6 hai6"}, // HKCanCor only
		{Traditional: "但是", Simplified: "但是", Jyutping: "daan6 si6"},  // wordfreq only
		{Traditional: "行", Simplified: "行", Jyutping: "hang4"},        // both: higher wins
		{Traditional: "嘢", Simplified: "嘢", Jyutping: "je5"},          // by reading
		{Traditional: "冇字", Simplified: "冇字", Jyutping: "mou5 zi6"},   // neither
	}
	if n := ApplyFrequencies(entries, wf, hk); n != 4 {
		t.Errorf("scored %d entries, want 4", n)
	}
	if entries[0].Frequency < 8 || entries[1].Frequency != 6.1 || entries[3].Frequency == 0 || entries[4].Frequency != 0 {
		t.Errorf("frequencies = %v %v %v %v", entries[0].Frequency, entries[1].Frequency, entries[3].Frequency, entries[4].Frequency)
	}
	if entries[2].Frequency <= 5.5 {
		t.Errorf("行 should take the higher HKCanCor score, got %v", entries[2].Frequency)
	}

	if _, err := LoadHKCanCor(writeHKCanCorZip(t, map[string]string{"utf8/x": "。/w/VQ1/\n"})); err == nil {
		t.Error("expected error for a corpus with no words")
	}
	if _, err := LoadHKCanCor(filepath.Join(t.TempDir(), "missing.zip")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseHKCanCorToken(t *testing.T) {
	for _, bad := range []string{"", "<sent>", "A:", "。/w/VQ1/", "/n/x/", "字/n//", "<x>/n/a1/"} {
		if _, _, _, ok := parseHKCanCorToken(bad); ok {
			t.Errorf("%q should be rejected", bad)
		}
	}
	if w, jp, tag, ok := parseHKCanCorToken("\t\t官司/n/Gun1si1/"); !ok || w != "官司" || jp != "gun1si1" || tag != "n" {
		t.Errorf("got %q %q %q %v", w, jp, tag, ok)
	}
}

func TestApplyFrequenciesSharedSimplified(t *testing.T) {
	wf := WordfreqTable{"于": 7.5, "去": 6.8}
	entries := []dictionary.Entry{
		{Traditional: "於", Simplified: "于", Jyutping: "jyu1"}, // the common word behind the score
		{Traditional: "于", Simplified: "于", Jyutping: "jyu1"}, // rare; shares the simplified key
		{Traditional: "去", Simplified: "去", Jyutping: "heoi3"},
	}
	ApplyFrequencies(entries, wf, CantoneseFreq{})
	if entries[0].Frequency != 7.5 || entries[1].Frequency != 6 || entries[2].Frequency != 6.8 {
		t.Errorf("got 於=%v 于=%v 去=%v, want 7.5 6 6.8", entries[0].Frequency, entries[1].Frequency, entries[2].Frequency)
	}
}
