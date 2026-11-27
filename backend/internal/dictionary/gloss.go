package dictionary

import (
	"regexp"
	"strings"
)

// Search matching works on two normalised forms of an entry's English glosses,
// computed once at insert time and stored next to the entry:
//
//   - gloss keys: each separate meaning, cleaned up, stored as "|key1|key2|".
//     "(conjunction) however; but;" → |however|but|, "to eat" → |to eat|eat|.
//     A query equal to a key is an exact match.
//   - gloss words: every word outside brackets, stored as " w1 w2 ... ".
//     A query word found here is a whole-word match ("but", not "butt").

var (
	// Bracketed labels and asides: (conjunction), (of water), [colloquial], （書面語）.
	bracketRe = regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]|（[^）]*）`)
	// Anything that is not a letter, digit or apostrophe separates words.
	nonWordRe = regexp.MustCompile(`[^\p{L}\p{N}']+`)
	// Sense numbering inside a gloss, as in CC-Canto's "1. still; 2. calm 1. or". Each number
	// starts a new meaning.
	senseNumberRe = regexp.MustCompile(`(^|\s)\d{1,2}[.)](\s|$)`)
)

const keyTrimChars = " .!?…'\"“”‘’:"

func isMeaningSeparator(r rune) bool {
	switch r {
	case ';', ',', '/', '；', '，':
		return true
	}
	return false
}

// glossKeys returns the entry's meanings as "|key1|key2|", or "" if it has none.
func glossKeys(english []string) string {
	keys := MeaningKeys(english)
	if len(keys) == 0 {
		return ""
	}
	return "|" + strings.Join(keys, "|") + "|"
}

// MeaningKeys splits glosses into cleaned-up meanings, in order: brackets removed, lists split
// on ; , / and sense numbers, lowercased, and "to eat" also listed as "eat".
func MeaningKeys(english []string) []string {
	seen := make(map[string]bool)
	var keys []string
	add := func(k string) {
		if k == "" || seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	for _, gloss := range english {
		cleaned := bracketRe.ReplaceAllString(gloss, " ")
		cleaned = senseNumberRe.ReplaceAllString(cleaned, ";")
		for _, part := range strings.FieldsFunc(cleaned, isMeaningSeparator) {
			key := strings.Trim(normalizeForExactMatch(part), keyTrimChars)
			add(key)
			if rest, ok := strings.CutPrefix(key, "to "); ok {
				add(strings.TrimSpace(rest))
			}
		}
	}
	return keys
}

// glossWords returns the lowercase words outside brackets as " w1 w2 ... ".
func glossWords(english []string) string {
	text := bracketRe.ReplaceAllString(strings.Join(english, " "), " ")
	text = senseNumberRe.ReplaceAllString(text, " ")
	words := queryWords(text)
	if len(words) == 0 {
		return ""
	}
	return " " + strings.Join(words, " ") + " "
}

// queryWords splits text into lowercase words the same way glossWords does.
func queryWords(text string) []string {
	return strings.Fields(nonWordRe.ReplaceAllString(strings.ToLower(text), " "))
}

// exactKey normalises a search query the same way a gloss key is normalised.
func exactKey(query string) string {
	return strings.Trim(normalizeForExactMatch(query), keyTrimChars)
}
