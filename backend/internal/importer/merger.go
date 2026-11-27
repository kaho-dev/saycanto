package importer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// MergeEntries combines base CC-Canto entries with a curated custom overlay.
//
// Entries are matched on their Traditional+Jyutping key. A curated overlay entry
// replaces every base entry with the same key (CC-Canto often lists one word on
// several lines); it takes the position of the first of them. Overlay entries
// with no base match are appended in order.
//
// IDs in the result are unique: curated IDs are kept as written (they appear in
// URLs), and any base entry whose ID collides is renumbered with the next free
// "-N" suffix.
func MergeEntries(base, overlay []dictionary.Entry) []dictionary.Entry {
	overlayByKey := make(map[string]dictionary.Entry, len(overlay))
	for _, e := range overlay {
		overlayByKey[mergeKey(e)] = e
	}

	merged := make([]dictionary.Entry, 0, len(base)+len(overlay))
	curated := make([]bool, 0, len(base)+len(overlay))
	placed := make(map[string]bool, len(overlay))

	for _, e := range base {
		k := mergeKey(e)
		if ov, ok := overlayByKey[k]; ok {
			if !placed[k] {
				merged = append(merged, ov)
				curated = append(curated, true)
				placed[k] = true
			}
			continue
		}
		merged = append(merged, e)
		curated = append(curated, false)
	}

	for _, e := range overlay {
		k := mergeKey(e)
		if placed[k] {
			continue
		}
		merged = append(merged, overlayByKey[k])
		curated = append(curated, true)
		placed[k] = true
	}

	return ensureUniqueIDs(merged, curated)
}

// ValidateOverlay checks curated entries for problems that would break the build.
func ValidateOverlay(entries []dictionary.Entry) error {
	seen := make(map[string]int, len(entries))
	for i, e := range entries {
		if strings.TrimSpace(e.ID) == "" {
			return fmt.Errorf("custom entry #%d (%q): missing id", i+1, e.Traditional)
		}
		if prev, dup := seen[e.ID]; dup {
			return fmt.Errorf("custom entries #%d and #%d share id %q", prev+1, i+1, e.ID)
		}
		seen[e.ID] = i
	}
	return nil
}

func mergeKey(e dictionary.Entry) string {
	return e.Traditional + "|" + e.Jyutping
}

func ensureUniqueIDs(entries []dictionary.Entry, curated []bool) []dictionary.Entry {
	used := make(map[string]bool, len(entries))
	for i, e := range entries {
		if curated[i] {
			used[e.ID] = true
		}
	}
	for i := range entries {
		if curated[i] {
			continue
		}
		id := entries[i].ID
		if used[id] {
			stem := stripNumericSuffix(id)
			for n := 2; ; n++ {
				candidate := fmt.Sprintf("%s-%d", stem, n)
				if !used[candidate] {
					id = candidate
					break
				}
			}
			entries[i].ID = id
		}
		used[id] = true
	}
	return entries
}

// stripNumericSuffix removes a trailing "-N" dedupe suffix. Jyutping syllables
// always contain letters, so an all-digit final segment can only be a suffix.
func stripNumericSuffix(id string) string {
	idx := strings.LastIndex(id, "-")
	if idx <= 0 {
		return id
	}
	if _, err := strconv.Atoi(id[idx+1:]); err != nil {
		return id
	}
	return id[:idx]
}
