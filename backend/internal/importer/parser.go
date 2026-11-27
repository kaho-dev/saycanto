package importer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

var ccCantoLineRe = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\[([^\]]*)\]\s+\{([^\}]+)\}\s+/(.+)/$`)

// trailingCommentRe matches a source note after the last definition slash,
// e.g. `/a surname/ # adapted from cc-cedict` (about a quarter of CC-Canto lines).
var trailingCommentRe = regexp.MustCompile(`^(.*/)\s+#[^/]*$`)

// ParseCCCanto parses CC-Canto formatted lines into dictionary entries.
func ParseCCCanto(r io.Reader) ([]dictionary.Entry, error) {
	scanner := bufio.NewScanner(r)
	var entries []dictionary.Entry
	seen := make(map[string]int)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if c := trailingCommentRe.FindStringSubmatch(line); c != nil {
			line = c[1]
		}
		m := ccCantoLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		trad, simp, jyutping, rawDefs := m[1], m[2], strings.TrimSpace(m[4]), m[5]
		defs := splitNonEmpty(rawDefs, "/")
		if len(defs) == 0 {
			continue
		}

		baseID := slugify(jyutping)
		if baseID == "" {
			baseID = trad
		}
		seen[baseID]++
		id := baseID
		if seen[baseID] > 1 {
			id = fmt.Sprintf("%s-%d", baseID, seen[baseID])
		}

		entries = append(entries, dictionary.Entry{
			ID:          id,
			English:     defs,
			Jyutping:    jyutping,
			Traditional: trad,
			Simplified:  simp,
			Register:    "spoken",
			Source:      dictionary.SourceCCCanto,
		})
	}
	return entries, scanner.Err()
}

// ParseCustomJSON parses a JSON array of curated dictionary entries.
func ParseCustomJSON(r io.Reader) ([]dictionary.Entry, error) {
	var entries []dictionary.Entry
	if err := json.NewDecoder(r).Decode(&entries); err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Source = dictionary.SourceCustom
	}
	return entries, nil
}

func splitNonEmpty(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func slugify(jyutping string) string {
	fields := strings.Fields(strings.ToLower(jyutping))
	return strings.Join(fields, "-")
}
