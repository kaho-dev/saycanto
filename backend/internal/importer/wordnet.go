package importer

import (
	"archive/zip"
	"bufio"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// English word types from Open English WordNet (CC BY 4.0), used only for entries that got no
// type from curated data, labels, HKCanCor or meaning patterns. WordNet says how many senses
// an English word has as a noun, adjective or adverb; this edition has no usage counts.
//
// Rules, to stay on the safe side:
//   - only an entry's first wordnetMeanings meanings are looked at, and only short ones;
//   - a meaning written "to …" is a verb (already handled), so for bare meanings WordNet's verb
//     senses are ignored: CC-CEDICT writes "water" for the noun and "to water" for the verb;
//   - a type is given only if it has at least wordnetDominance of the word's remaining senses
//     ("water": noun 6/6 → noun; "light": adjective 25/41 → no type).

const (
	wordnetMeanings  = 3
	wordnetMaxWords  = 3
	wordnetDominance = 0.7
)

// WordNet maps an English lemma ("ice cream") to its sense count per word type.
type WordNet struct {
	senses map[string]map[string]int // lemma → "noun"/"adjective"/"adverb" → synsets
	plural map[string]string         // irregular plural → lemma (noun.exc)
}

var wordnetIndexTypes = map[string]string{"index.noun": "noun", "index.adj": "adjective", "index.adv": "adverb"}

// LoadWordNet reads index.noun / index.adj / index.adv and noun.exc from a WNDB-format zip.
func LoadWordNet(zipPath string) (*WordNet, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	wn := &WordNet{senses: map[string]map[string]int{}, plural: map[string]string{}}
	found := 0
	for _, f := range zr.File {
		base := path.Base(f.Name)
		wordType, isIndex := wordnetIndexTypes[base]
		if !isIndex && base != "noun.exc" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(rc)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || line[0] == ' ' { // licence header lines start with spaces
				continue
			}
			fields := strings.Fields(line)
			if isIndex {
				// lemma pos synset_cnt ...
				if len(fields) < 3 {
					continue
				}
				n, err := strconv.Atoi(fields[2])
				if err != nil || n <= 0 {
					continue
				}
				lemma := strings.ReplaceAll(strings.ToLower(fields[0]), "_", " ")
				if wn.senses[lemma] == nil {
					wn.senses[lemma] = map[string]int{}
				}
				wn.senses[lemma][wordType] += n
			} else if len(fields) >= 2 {
				wn.plural[strings.ReplaceAll(fields[0], "_", " ")] = strings.ReplaceAll(fields[1], "_", " ")
			}
		}
		_ = rc.Close()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		if isIndex {
			found++
		}
	}
	if found == 0 {
		return nil, fmt.Errorf("%s: no WordNet index files found", zipPath)
	}
	return wn, nil
}

// TypeOf returns the dominant noun/adjective/adverb type of a bare English meaning, or "".
func (wn *WordNet) TypeOf(meaning string) string {
	m := strings.TrimSpace(strings.ToLower(meaning))
	for _, article := range []string{"a ", "an ", "the "} {
		m = strings.TrimPrefix(m, article)
	}
	if m == "" || len(strings.Fields(m)) > wordnetMaxWords {
		return ""
	}
	counts := wn.senses[m]
	if counts == nil {
		// Plurals: "drinks" → drink, "mice" → mouse.
		if lemma, ok := wn.plural[m]; ok {
			counts = map[string]int{"noun": wn.senses[lemma]["noun"]}
		} else if lemma := singular(m); lemma != m && wn.senses[lemma]["noun"] > 0 {
			counts = map[string]int{"noun": wn.senses[lemma]["noun"]}
		}
	}
	total := 0
	best, bestN := "", 0
	for _, t := range []string{"noun", "adjective", "adverb"} {
		n := counts[t]
		total += n
		if n > bestN {
			best, bestN = t, n
		}
	}
	if total == 0 || float64(bestN)/float64(total) < wordnetDominance {
		return ""
	}
	return best
}

func singular(w string) string {
	switch {
	case strings.HasSuffix(w, "ies") && len(w) > 4:
		return w[:len(w)-3] + "y"
	case strings.HasSuffix(w, "ses") || strings.HasSuffix(w, "xes") || strings.HasSuffix(w, "ches") || strings.HasSuffix(w, "shes"):
		return w[:len(w)-2]
	case strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && len(w) > 3:
		return w[:len(w)-1]
	}
	return w
}

// InferWordTypesFromWordNet types the still-untyped entries from their first meanings and
// returns how many entries it typed.
func InferWordTypesFromWordNet(entries []dictionary.Entry, wn *WordNet) int {
	if wn == nil {
		return 0
	}
	typed := 0
	for i := range entries {
		e := &entries[i]
		if len(e.WordTypes) > 0 {
			continue
		}
		keys := dictionary.MeaningKeys(e.English)
		verbStems := map[string]bool{}
		for _, k := range keys {
			if rest, ok := strings.CutPrefix(k, "to "); ok {
				verbStems[strings.TrimSpace(rest)] = true
			}
		}
		set := map[string]bool{}
		looked := 0
		for _, k := range keys {
			if looked == wordnetMeanings {
				break
			}
			if strings.HasPrefix(k, "to ") {
				looked++ // a verb meaning counts as one of the first meanings
				continue
			}
			if verbStems[k] {
				continue // the "eat" copy of "to eat"
			}
			looked++
			if t := wn.TypeOf(k); t != "" {
				set[t] = true
			}
		}
		for _, t := range dictionary.WordTypes {
			if set[t] {
				e.WordTypes = append(e.WordTypes, t)
			}
		}
		if len(e.WordTypes) > 0 {
			typed++
		}
	}
	return typed
}
