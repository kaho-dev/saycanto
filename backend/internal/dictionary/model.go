package dictionary

import "errors"

// ErrNotFound is returned when a dictionary entry cannot be found.
var ErrNotFound = errors.New("dictionary entry not found")

// ErrInvalidInput is returned when query or ID parameters are invalid.
var ErrInvalidInput = errors.New("invalid input")

// Example represents a sample Cantonese sentence with Jyutping and English.
type Example struct {
	Jyutping  string `json:"jyutping"`
	Cantonese string `json:"cantonese"`
	English   string `json:"english"`
}

// Entry represents a single English-to-Cantonese dictionary entry.
type Entry struct {
	ID           string    `json:"id"`
	English      []string  `json:"english"`
	Jyutping     string    `json:"jyutping"`
	Traditional  string    `json:"traditional"`
	Simplified   string    `json:"simplified"`
	PartOfSpeech string    `json:"partOfSpeech,omitempty"`
	Register     string    `json:"register,omitempty"` // spoken, formal, both
	Classifier   string    `json:"classifier,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	Examples     []Example `json:"examples,omitempty"`
	Source       string    `json:"source,omitempty"` // SourceCustom, SourceCCCanto or SourceCEDICT; set by the importer
	// Frequency is how common the word is on the Zipf scale (0 = unknown, ~5 = everyday).
	// Set by the importer and used only for ranking, so it is not part of the API.
	Frequency float64 `json:"-"`
	// WordTypes lists the entry's parts of speech (values from WordTypes), worked out by the importer.
	// Empty when unknown.
	WordTypes []string `json:"wordTypes,omitempty"`
}

// WordTypes are the parts of speech an entry can be tagged with, in display order.
var WordTypes = []string{
	"noun", "verb", "adjective", "adverb", "pronoun", "preposition",
	"conjunction", "classifier", "particle", "interjection", "number", "expression",
}

// IsWordType reports whether t is one of WordTypes.
func IsWordType(t string) bool {
	for _, w := range WordTypes {
		if w == t {
			return true
		}
	}
	return false
}

// Entry sources. Curated (custom) entries rank above bulk CC-Canto entries.
const (
	SourceCustom  = "custom"
	SourceCCCanto = "cc-canto"
	SourceCEDICT  = "cc-cedict"
)

// SearchQuery encapsulates parameters for searching dictionary entries.
type SearchQuery struct {
	Query    string
	Register string
	WordType string // one of WordTypes, or "" for all
	Limit    int
}

// SearchResult is the API response model for search requests.
type SearchResult struct {
	Query string `json:"query"`
	// Total is how many entries match (with the word-type filter, if any), not just this page.
	Total   int     `json:"total"`
	Entries []Entry `json:"entries"`
	// TypeCounts counts the matches per word type, ignoring the word-type filter, so the
	// filter can show every option. Untyped counts matches with no known word type.
	TypeCounts map[string]int `json:"typeCounts"`
	Untyped    int            `json:"untyped"`
}

// TypeCounts is the per-word-type breakdown of a search's matches.
type TypeCounts struct {
	Total   int
	ByType  map[string]int
	Untyped int
}

// ListQuery encapsulates pagination and filter parameters for listing dictionary entries.
type ListQuery struct {
	Register string
	WordType string
	Limit    int
	Offset   int
}

// ListResult is the API response model for paginated word list requests.
type ListResult struct {
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	Entries []Entry `json:"entries"`
}
