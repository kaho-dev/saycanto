package importer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

const sampleReadings = `# CC-CEDICT Cantonese Readings
通常 通常 [tong1 chang2] {tung1 soeng4}
凄切 凄切 [qi1qie4] {cai1 cit3}
行 行 [hang2] {hong4}
行 行 [xing2] {hang4}
行 行 [xing2] {hang4}
平時 平时 [ping2 shi2] {ping4 si4}
綠 绿 [lu:4] {luk6}
發 发 [fa1] {faat3}
malformed
`

const sampleCEDICT = `# CC-CEDICT
#! version=1
通常 通常 [tong1 chang2] /regular; usual; normal/usually; normally/
凄切 凄切 [[qi1qie4]] /mournful/
行 行 [xing2] /to walk/to go/capable/
平時 平时 [ping2 shi2] /ordinarily; in normal times/
綠 绿 [[lü4]] /green/
發 发 [fa1] /variant of 髮|发[fa4]/see 頭髮|头发[tou2 fa5]/CL:個|个[ge4]/
無讀音 无读音 [wu2 du2 yin1] /no Cantonese reading here/
`

func TestParseCantoReadings(t *testing.T) {
	readings, err := ParseCantoReadings(strings.NewReader(sampleReadings))
	if err != nil {
		t.Fatal(err)
	}
	if got := readings[readingKey("行", "行", "xing2")]; !reflect.DeepEqual(got, []string{"hang4"}) {
		t.Errorf("duplicate readings not collapsed: %v", got)
	}
	if got := readings[readingKey("凄切", "凄切", "qi1 qie4")]; !reflect.DeepEqual(got, []string{"cai1 cit3"}) {
		t.Errorf("pinyin spacing not normalised: %v", got)
	}
}

func TestParseCEDICT(t *testing.T) {
	readings, _ := ParseCantoReadings(strings.NewReader(sampleReadings))
	entries, stats, err := ParseCEDICT(strings.NewReader(sampleCEDICT), readings)
	if err != nil {
		t.Fatal(err)
	}
	want := CEDICTStats{Lines: 7, Entries: 5, NoReading: 1, OnlyReference: 1}
	if stats != want {
		t.Errorf("stats = %+v, want %+v", stats, want)
	}
	byTrad := map[string]dictionary.Entry{}
	for _, e := range entries {
		byTrad[e.Traditional] = e
	}
	tung := byTrad["通常"]
	if tung.Jyutping != "tung1 soeng4" || tung.ID != "tung1-soeng4" || tung.Source != dictionary.SourceCEDICT || tung.Register != "" {
		t.Errorf("通常 = %+v", tung)
	}
	if !reflect.DeepEqual(tung.English, []string{"regular; usual; normal", "usually; normally"}) {
		t.Errorf("通常 senses = %q", tung.English)
	}
	if byTrad["凄切"].Jyutping != "cai1 cit3" || byTrad["綠"].Jyutping != "luk6" {
		t.Errorf("v2 pinyin not matched: %+v / %+v", byTrad["凄切"], byTrad["綠"])
	}
	if _, ok := byTrad["發"]; ok {
		t.Error("cross-reference-only entry should be skipped")
	}
}

func TestCombineBaseMergesMeanings(t *testing.T) {
	primary := []dictionary.Entry{{ID: "ping4-si4", Traditional: "平時", Jyutping: "ping4 si4", English: []string{"(noun) spare time"}, Source: dictionary.SourceCCCanto}}
	secondary := []dictionary.Entry{
		{ID: "ping4-si4", Traditional: "平時", Jyutping: "ping4 si4", English: []string{"ordinarily; in normal times", "(Noun) Spare time"}},
		{ID: "tung1-soeng4", Traditional: "通常", Jyutping: "tung1 soeng4", English: []string{"usually"}},
	}
	got := CombineBase(primary, secondary)
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	if !reflect.DeepEqual(got[0].English, []string{"(noun) spare time", "ordinarily; in normal times"}) || got[0].Source != dictionary.SourceCCCanto {
		t.Errorf("merged 平時 = %+v", got[0])
	}
	if !reflect.DeepEqual(primary[0].English, []string{"(noun) spare time"}) {
		t.Error("CombineBase must not modify its input")
	}
}

func TestOpenMaybeZip(t *testing.T) {
	dir := t.TempDir()
	txt := filepath.Join(dir, "cedict.txt")
	if err := os.WriteFile(txt, []byte("plain"), 0o644); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(dir, "cedict.zip")
	zf, _ := os.Create(zipPath)
	zw := zip.NewWriter(zf)
	w, _ := zw.Create("cedict_ts.u8")
	_, _ = w.Write([]byte("zipped"))
	_ = zw.Close()
	_ = zf.Close()

	emptyZip := filepath.Join(dir, "empty.zip")
	ef, _ := os.Create(emptyZip)
	ew := zip.NewWriter(ef)
	_, _ = ew.Create("readme.md")
	_ = ew.Close()
	_ = ef.Close()

	for path, want := range map[string]string{txt: "plain", zipPath: "zipped"} {
		r, closeFn, err := OpenMaybeZip(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		b := new(strings.Builder)
		buf := make([]byte, 64)
		n, _ := r.Read(buf)
		b.Write(buf[:n])
		_ = closeFn()
		if b.String() != want {
			t.Errorf("%s: read %q, want %q", path, b.String(), want)
		}
	}
	if _, _, err := OpenMaybeZip(emptyZip); err == nil {
		t.Error("expected error for zip without a text file")
	}
	if _, _, err := OpenMaybeZip(filepath.Join(dir, "missing.txt")); err == nil {
		t.Error("expected error for missing file")
	}
	if _, _, err := OpenMaybeZip(filepath.Join(dir, "missing.zip")); err == nil {
		t.Error("expected error for missing zip")
	}
}
