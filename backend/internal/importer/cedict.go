package importer

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// CC-CEDICT is a Chinese–English dictionary with Mandarin (pinyin) readings only.
// Pleco's "CC-CEDICT Cantonese readings" file supplies Jyutping for its entries,
// keyed by the same traditional + simplified + pinyin triple:
//
//	CC-CEDICT: 通常 通常 [tong1 chang2] /regular; usual; normal/usually; normally/
//	readings:  通常 通常 [tong1 chang2] {tung1 soeng4}
//
// CC-CEDICT's newer "v2" syntax writes pinyin in double brackets without spaces
// ([[tong1chang2]]); both forms are accepted.

var (
	readingLineRe = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\[\[?([^\]]*)\]?\]\s+\{([^\}]+)\}`)
	cedictLineRe  = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\[\[?([^\]]*)\]?\]\s+/(.+)/\s*$`)
	// Senses that only point at another headword or carry Mandarin-only metadata.
	cedictSkipSenseRe = regexp.MustCompile(`^(?i)((old|archaic|ancient|variant|japanese|korean|erroneous|obscure)\s+)*variant of |^see (also )?\S+\[|^CL:|^Taiwan pr\.|^also pr\.|^abbr\. for `)
)

// CantoReadings maps a CC-CEDICT headword (traditional, simplified, pinyin) to its Jyutping readings.
type CantoReadings map[string][]string

func readingKey(trad, simp, pinyin string) string {
	return trad + "|" + simp + "|" + normalizePinyin(pinyin)
}

// normalizePinyin makes v1 ("tong1 chang2", "lu:4") and v2 ("tong1chang2", "lü4") pinyin comparable.
func normalizePinyin(p string) string {
	p = strings.ToLower(p)
	p = strings.NewReplacer("ü", "u:", "v", "u:").Replace(p)
	var b strings.Builder
	for _, r := range p {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ':' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ParseCantoReadings parses the CC-CEDICT Cantonese readings file.
func ParseCantoReadings(r io.Reader) (CantoReadings, error) {
	readings := make(CantoReadings)
	scanner := newLineScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := readingLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		jyutping := strings.Join(strings.Fields(strings.ToLower(m[4])), " ")
		key := readingKey(m[1], m[2], m[3])
		if !containsString(readings[key], jyutping) {
			readings[key] = append(readings[key], jyutping)
		}
	}
	return readings, scanner.Err()
}

// CEDICTStats reports what ParseCEDICT kept and skipped.
type CEDICTStats struct {
	Lines         int // dictionary lines read (not comments)
	Entries       int // entries produced (one per Jyutping reading)
	NoReading     int // lines skipped: no Cantonese reading available
	OnlyReference int // lines skipped: every sense was "variant of ...", "see ...", "CL:..." etc.
}

// ParseCEDICT parses CC-CEDICT and keeps only headwords that have a Cantonese reading.
// It never guesses a reading: a wrong tone is a wrong word.
func ParseCEDICT(r io.Reader, readings CantoReadings) ([]dictionary.Entry, CEDICTStats, error) {
	var entries []dictionary.Entry
	var stats CEDICTStats
	scanner := newLineScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := cedictLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		stats.Lines++

		trad, simp, pinyin := m[1], m[2], m[3]
		jyutpings := readings[readingKey(trad, simp, pinyin)]
		if len(jyutpings) == 0 {
			stats.NoReading++
			continue
		}

		var senses []string
		for _, s := range splitNonEmpty(m[4], "/") {
			if !cedictSkipSenseRe.MatchString(s) {
				senses = append(senses, s)
			}
		}
		if len(senses) == 0 {
			stats.OnlyReference++
			continue
		}

		for _, jp := range jyutpings {
			entries = append(entries, dictionary.Entry{
				ID:          slugify(jp),
				English:     append([]string(nil), senses...),
				Jyutping:    jp,
				Traditional: trad,
				Simplified:  simp,
				Source:      dictionary.SourceCEDICT,
			})
			stats.Entries++
		}
	}
	return entries, stats, scanner.Err()
}

// CombineBase adds secondary entries to primary ones. When both have the same
// traditional + Jyutping, the secondary meanings missing from the primary entry
// are appended to it (so 平時 keeps CC-Canto's "spare time" and gains CC-CEDICT's
// "ordinarily"); otherwise the secondary entry is added as its own entry.
// IDs are made unique later by MergeEntries.
func CombineBase(primary, secondary []dictionary.Entry) []dictionary.Entry {
	out := make([]dictionary.Entry, len(primary), len(primary)+len(secondary))
	copy(out, primary)
	byKey := make(map[string]int, len(out))
	for i, e := range out {
		if _, ok := byKey[mergeKey(e)]; !ok {
			byKey[mergeKey(e)] = i
		}
	}
	for _, e := range secondary {
		k := mergeKey(e)
		if i, ok := byKey[k]; ok {
			out[i].English = appendNewMeanings(out[i].English, e.English)
			continue
		}
		byKey[k] = len(out)
		out = append(out, e)
	}
	return out
}

func appendNewMeanings(have, add []string) []string {
	seen := make(map[string]bool, len(have))
	for _, h := range have {
		seen[strings.ToLower(strings.TrimSpace(h))] = true
	}
	result := append([]string(nil), have...)
	for _, a := range add {
		k := strings.ToLower(strings.TrimSpace(a))
		if !seen[k] {
			seen[k] = true
			result = append(result, a)
		}
	}
	return result
}

// OpenMaybeZip opens a text file, or the first .txt/.u8 file inside a .zip archive.
// The caller must call the returned close function.
func OpenMaybeZip(path string) (io.Reader, func() error, error) {
	if !strings.HasSuffix(strings.ToLower(path), ".zip") {
		f, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return f, f.Close, nil
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, err
	}
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".u8") {
			rc, err := f.Open()
			if err != nil {
				_ = zr.Close()
				return nil, nil, err
			}
			return rc, func() error { _ = rc.Close(); return zr.Close() }, nil
		}
	}
	_ = zr.Close()
	return nil, nil, fmt.Errorf("%s: no .txt or .u8 file inside", path)
}

// newLineScanner allows long dictionary lines (some CC-CEDICT entries exceed bufio's 64 KB default).
func newLineScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return s
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
