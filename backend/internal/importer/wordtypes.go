package importer

import (
	"regexp"
	"sort"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// Word types (parts of speech) are worked out from four signals, in this order of trust:
//
//  1. a curated entry's partOfSpeech;
//  2. labels at the start of a meaning: "(noun) ...", "(v.) ...", "(phrase; adverb; noun) ...",
//     "(idiom) ...";
//  3. HKCanCor's tags for that exact word + reading (a tag counts if it is at least 20% of the
//     word's uses there) — good for the small function words: 嘅, 咗, 但係, 佢;
//  4. meaning patterns: CC-CEDICT and CC-Canto write verbs as "to ...", and classifiers as
//     "classifier for ...".
//
// An entry can have several types (同: conjunction, preposition, adjective). Entries with no
// signal stay untyped; types are never guessed from the English alone beyond rule 4.

// hkcancorTagTypes maps HKCanCor (PKU-style) tags to word types. Tags not listed (x* code-mixed
// words, morphemes, prefixes/suffixes, punctuation) give no type.
var hkcancorTagTypes = map[string]string{
	"n": "noun", "nr": "noun", "ns": "noun", "nt": "noun", "nz": "noun", "nx": "noun", "Ng": "noun",
	"t": "noun", "f": "noun",
	"v": "verb", "v1": "verb", "vu": "verb", "vn": "verb", "vk": "verb", "Vg": "verb",
	"a": "adjective", "b": "adjective", "z": "adjective",
	"d": "adverb", "d1": "adverb",
	"r": "pronoun",
	"p": "preposition",
	"c": "conjunction",
	"q": "classifier",
	"y": "particle", "y1": "particle", "u": "particle",
	"e": "interjection", "o": "interjection",
	"m": "number",
	"i": "expression", "l": "expression", "l1": "expression",
}

// labelTypes maps words found in a leading "(...)" label to word types.
var labelTypes = map[string]string{
	"noun": "noun", "n": "noun", "n.": "noun",
	"verb": "verb", "v": "verb", "v.": "verb",
	"adjective": "adjective", "adj": "adjective", "adj.": "adjective",
	"adverb": "adverb", "adv": "adverb", "adv.": "adverb",
	"pronoun":     "pronoun",
	"preposition": "preposition", "prep": "preposition", "prep.": "preposition",
	"conjunction": "conjunction", "conj": "conjunction", "conj.": "conjunction",
	"classifier": "classifier", "measure word": "classifier",
	"particle":     "particle",
	"interjection": "interjection", "exclamation": "interjection",
	"numeral": "number", "number": "number",
	"idiom": "expression", "phrase": "expression", "expression": "expression", "set phrase": "expression",
}

var (
	leadingLabelRe = regexp.MustCompile(`^\(([^)]{1,40})\)`)
	verbGlossRe    = regexp.MustCompile(`(?i)^to [a-z]`)
	classifierRe   = regexp.MustCompile(`(?i)^classifier for `)
	particleRe     = regexp.MustCompile(`(?i)^((sentence|question|modal|aspect|final)[- ]?(final )?)?particle\b`)
	idiomRe        = regexp.MustCompile(`(?i)\((idiom|chengyu|proverb|saying)\)`)
)

// typeOrder fixes the order types are listed in.
var typeOrder = func() map[string]int {
	m := map[string]int{}
	for i, t := range dictionary.WordTypes {
		m[t] = i
	}
	return m
}()

// InferWordTypes fills WordTypes on every entry and returns how many entries got at least one.
func InferWordTypes(entries []dictionary.Entry, hk CantoneseFreq) int {
	typed := 0
	for i := range entries {
		e := &entries[i]
		set := map[string]bool{}
		add := func(t string) {
			if _, ok := typeOrder[t]; ok {
				set[t] = true
			}
		}

		if pos := strings.ToLower(strings.TrimSpace(e.PartOfSpeech)); pos != "" {
			add(labelTypes[pos])
		}
		// An idiom's "fig. to overdo it" sense describes the idiom; it does not make it a verb.
		idiom := false
		for _, sense := range e.English {
			if idiomRe.MatchString(sense) {
				idiom = true
				add("expression")
			}
		}
		for _, sense := range e.English {
			for _, part := range strings.Split(sense, "/") {
				part = strings.TrimSpace(part)
				if m := leadingLabelRe.FindStringSubmatch(part); m != nil {
					for _, label := range strings.FieldsFunc(m[1], func(r rune) bool { return r == ';' || r == ',' || r == '/' }) {
						add(labelTypes[strings.ToLower(strings.TrimSpace(label))])
					}
					part = strings.TrimSpace(part[len(m[0]):])
				}
				// A meaning that is only a label word: "verb; update bank passbook", "adj; faithful".
				for _, piece := range strings.FieldsFunc(part, func(r rune) bool { return r == ';' || r == ',' }) {
					if t, ok := labelTypes[strings.ToLower(strings.TrimSpace(piece))]; ok {
						add(t)
					}
				}
				if verbGlossRe.MatchString(part) && !idiom {
					add("verb")
				}
				if particleRe.MatchString(part) {
					add("particle")
				}
				if classifierRe.MatchString(part) {
					add("classifier")
				}
			}
		}
		for _, tag := range hk.Tags[e.Traditional+"|"+compactJyutping(e.Jyutping)] {
			add(hkcancorTagTypes[tag])
		}

		e.WordTypes = nil
		for t := range set {
			e.WordTypes = append(e.WordTypes, t)
		}
		sort.Slice(e.WordTypes, func(a, b int) bool { return typeOrder[e.WordTypes[a]] < typeOrder[e.WordTypes[b]] })
		if len(e.WordTypes) > 0 {
			typed++
		}
	}
	return typed
}
