package importer

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// Word frequencies, on the Zipf scale (log10 of uses per billion words: ~7 for 的,
// ~5 for an everyday word, ~3 for a rare one, 0 = unknown), from two sources:
//
//   - wordfreq's Chinese list (large_zh.msgpack.gz, CC BY-SA 4.0): general, mostly
//     Mandarin usage. Keys are "oversimplified" Chinese, so it is looked up by the
//     entry's Simplified form.
//   - HKCanCor (Hong Kong Cantonese Corpus, CC BY 4.0): ~154k words of transcribed
//     Hong Kong speech, tagged with Jyutping, so Cantonese words like 但係 or 嘢 get a
//     score too and are not outranked by their Mandarin equivalents.
//
// An entry gets the higher of the two scores.

// WordfreqTable maps a (simplified) word to its Zipf frequency.
type WordfreqTable map[string]float64

// LoadWordfreq reads a wordfreq "cBpack" file: gzipped msgpack holding
// [header, bucket0, bucket1, ...], where every word in bucket i has a frequency
// of -i centibels, i.e. Zipf = 9 - i/100.
func LoadWordfreq(r io.Reader) (WordfreqTable, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	v, err := decodeMsgpack(bufio.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("wordfreq: %w", err)
	}
	top, ok := v.([]any)
	if !ok || len(top) == 0 {
		return nil, errors.New("wordfreq: expected a list")
	}
	header, ok := top[0].(map[string]any)
	if !ok || header["format"] != "cB" {
		return nil, errors.New("wordfreq: not a cB-format file")
	}
	table := make(WordfreqTable)
	for i, bucket := range top[1:] {
		words, ok := bucket.([]any)
		if !ok {
			return nil, fmt.Errorf("wordfreq: bucket %d is not a list", i)
		}
		zipf := 9 - float64(i)/100
		for _, w := range words {
			if s, ok := w.(string); ok {
				if _, seen := table[s]; !seen {
					table[s] = zipf
				}
			}
		}
	}
	return table, nil
}

// CantoneseFreq holds HKCanCor frequencies by word + Jyutping. Matching on the reading
// as well as the characters keeps 噉 gam2 ("like this", very common) from lending its
// score to 噉 daam6 ("to eat", rare).
type CantoneseFreq struct {
	ByReading map[string]float64  // "唔該|m4goi1"
	Tags      map[string][]string // "唔該|m4goi1" → POS tags covering at least 20% of its uses
	Words     int                 // total word tokens counted
}

// LoadHKCanCor counts words in the HKCanCor zip (utf8/* transcript files with
// lines like "官司/n/gun1si1/"). Punctuation and words seen only once are ignored.
func LoadHKCanCor(path string) (CantoneseFreq, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return CantoneseFreq{}, err
	}
	defer zr.Close()

	byReading := map[string]int{}
	tagCounts := map[string]map[string]int{}
	total := 0
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "utf8/") || f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return CantoneseFreq{}, err
		}
		scanner := newLineScanner(rc)
		for scanner.Scan() {
			word, jp, tag, ok := parseHKCanCorToken(scanner.Text())
			if !ok {
				continue
			}
			total++
			key := word + "|" + jp
			byReading[key]++
			if tagCounts[key] == nil {
				tagCounts[key] = map[string]int{}
			}
			tagCounts[key][tag]++
		}
		_ = rc.Close()
		if err := scanner.Err(); err != nil {
			return CantoneseFreq{}, err
		}
	}
	if total == 0 {
		return CantoneseFreq{}, errors.New("hkcancor: no words found")
	}

	zipf := func(n int) float64 { return math.Log10(float64(n) / float64(total) * 1e9) }
	out := CantoneseFreq{ByReading: map[string]float64{}, Tags: map[string][]string{}, Words: total}
	for k, n := range byReading {
		if n >= 2 {
			out.ByReading[k] = zipf(n)
		}
		for tag, c := range tagCounts[k] {
			if c*5 >= n { // at least 20% of this word's uses
				out.Tags[k] = append(out.Tags[k], tag)
			}
		}
		sort.Strings(out.Tags[k])
	}
	return out, nil
}

// parseHKCanCorToken reads "官司/n/gun1si1/" → ("官司", "gun1si1", "n"). Punctuation (POS "w")
// and lines that are not word tokens are rejected.
func parseHKCanCorToken(line string) (word, jyutping, tag string, ok bool) {
	parts := strings.Split(strings.TrimSpace(line), "/")
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[1] == "w" || parts[2] == "" {
		return "", "", "", false
	}
	if strings.ContainsAny(parts[0], "<>") {
		return "", "", "", false
	}
	return parts[0], compactJyutping(parts[2]), parts[1], true
}

func compactJyutping(jp string) string {
	return strings.ReplaceAll(strings.ToLower(jp), " ", "")
}

// sharedSimplifiedPenalty lowers the wordfreq score of a headword whose traditional form
// equals its simplified form when other traditional forms simplify to the same characters.
// wordfreq only knows simplified text, so 于 ("to go", rare) would otherwise inherit the
// score of 於 (very common), whose simplified form is also 于.
const sharedSimplifiedPenalty = 1.5

// ApplyFrequencies sets each entry's Frequency to the higher of its wordfreq and HKCanCor scores.
// It returns how many entries received a non-zero score.
func ApplyFrequencies(entries []dictionary.Entry, wf WordfreqTable, hk CantoneseFreq) int {
	// Which simplified forms are shared by a traditional form that differs from them?
	sharedWithOther := map[string]bool{}
	for _, e := range entries {
		if e.Traditional != e.Simplified {
			sharedWithOther[e.Simplified] = true
		}
	}

	scored := 0
	for i := range entries {
		e := &entries[i]
		best := wf[e.Simplified]
		if best > 0 && e.Traditional == e.Simplified && sharedWithOther[e.Simplified] {
			best = math.Max(0, best-sharedSimplifiedPenalty)
		}
		if v := hk.ByReading[e.Traditional+"|"+compactJyutping(e.Jyutping)]; v > best {
			best = v
		}
		e.Frequency = math.Round(best*100) / 100
		if e.Frequency > 0 {
			scored++
		}
	}
	return scored
}

// decodeMsgpack decodes the subset of MessagePack used by wordfreq files:
// nil, booleans, integers, strings, arrays and string-keyed maps.
func decodeMsgpack(r *bufio.Reader) (any, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch {
	case b <= 0x7f:
		return int64(b), nil
	case b >= 0xe0:
		return int64(int8(b)), nil
	case b >= 0xa0 && b <= 0xbf:
		return readStr(r, int(b&0x1f))
	case b >= 0x90 && b <= 0x9f:
		return readArray(r, int(b&0x0f))
	case b >= 0x80 && b <= 0x8f:
		return readMap(r, int(b&0x0f))
	}
	switch b {
	case 0xc0:
		return nil, nil
	case 0xc2:
		return false, nil
	case 0xc3:
		return true, nil
	case 0xcc, 0xd0:
		n, err := readUint(r, 1)
		if b == 0xd0 {
			return int64(int8(n)), err
		}
		return int64(n), err
	case 0xcd, 0xd1:
		n, err := readUint(r, 2)
		if b == 0xd1 {
			return int64(int16(n)), err
		}
		return int64(n), err
	case 0xce, 0xd2:
		n, err := readUint(r, 4)
		if b == 0xd2 {
			return int64(int32(n)), err
		}
		return int64(n), err
	case 0xd9, 0xda, 0xdb:
		n, err := readUint(r, map[byte]int{0xd9: 1, 0xda: 2, 0xdb: 4}[b])
		if err != nil {
			return nil, err
		}
		return readStr(r, int(n))
	case 0xdc, 0xdd:
		n, err := readUint(r, map[byte]int{0xdc: 2, 0xdd: 4}[b])
		if err != nil {
			return nil, err
		}
		return readArray(r, int(n))
	case 0xde, 0xdf:
		n, err := readUint(r, map[byte]int{0xde: 2, 0xdf: 4}[b])
		if err != nil {
			return nil, err
		}
		return readMap(r, int(n))
	}
	return nil, fmt.Errorf("unsupported msgpack type 0x%02x", b)
}

func readUint(r *bufio.Reader, size int) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(r, buf[8-size:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}

func readStr(r *bufio.Reader, n int) (string, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readArray(r *bufio.Reader, n int) ([]any, error) {
	out := make([]any, n)
	for i := range out {
		v, err := decodeMsgpack(r)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func readMap(r *bufio.Reader, n int) (map[string]any, error) {
	out := make(map[string]any, n)
	for i := 0; i < n; i++ {
		k, err := decodeMsgpack(r)
		if err != nil {
			return nil, err
		}
		v, err := decodeMsgpack(r)
		if err != nil {
			return nil, err
		}
		ks, ok := k.(string)
		if !ok {
			return nil, errors.New("msgpack map key is not a string")
		}
		out[ks] = v
	}
	return out, nil
}
