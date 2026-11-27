package dictionary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// senseRankPenalty is the Zipf-point cost per position the matching meaning sits down an entry's list.
const senseRankPenalty = "0.3"

const schemaSQL = `
CREATE TABLE IF NOT EXISTS entries (
	id TEXT PRIMARY KEY,
	english_json TEXT NOT NULL,
	jyutping TEXT NOT NULL,
	traditional TEXT NOT NULL,
	simplified TEXT NOT NULL,
	part_of_speech TEXT NOT NULL DEFAULT '',
	register TEXT NOT NULL DEFAULT '',
	classifier TEXT NOT NULL DEFAULT '',
	notes TEXT NOT NULL DEFAULT '',
	examples_json TEXT NOT NULL DEFAULT '[]',
	source TEXT NOT NULL DEFAULT '',
	gloss_keys TEXT NOT NULL DEFAULT '',
	gloss_words TEXT NOT NULL DEFAULT '',
	freq REAL NOT NULL DEFAULT 0,
	word_types TEXT NOT NULL DEFAULT ''
);

CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
	id UNINDEXED,
	english_text,
	jyutping,
	traditional,
	simplified,
	tokenize = 'unicode61'
);
`

// SQLiteRepository implements Repository using an embedded SQLite FTS5 database.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a SQLiteRepository backed by the given *sql.DB.
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// InitSchema creates the dictionary tables and FTS5 virtual table if needed,
// and adds columns introduced after a database was first built.
func InitSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	// Columns added after the first release. gloss_keys / gloss_words are filled by
	// InsertEntries, so an upgraded database needs a re-import for best search ranking.
	for _, col := range []struct{ name, def string }{
		{"source", "TEXT NOT NULL DEFAULT ''"},
		{"gloss_keys", "TEXT NOT NULL DEFAULT ''"},
		{"gloss_words", "TEXT NOT NULL DEFAULT ''"},
		{"freq", "REAL NOT NULL DEFAULT 0"},
		{"word_types", "TEXT NOT NULL DEFAULT ''"},
	} {
		if err := addColumnIfMissing(db, "entries", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
	return err
}

// InsertEntries inserts or replaces entries and rebuilds their FTS5 index rows.
func InsertEntries(ctx context.Context, db *sql.DB, entries []Entry) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM entries_fts; DELETE FROM entries;`); err != nil {
		return err
	}

	stmtEntry, err := tx.PrepareContext(ctx, `
		INSERT INTO entries (id, english_json, jyutping, traditional, simplified, part_of_speech, register, classifier, notes, examples_json, source, gloss_keys, gloss_words, freq, word_types)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtEntry.Close()

	stmtFTS, err := tx.PrepareContext(ctx, `
		INSERT INTO entries_fts (id, english_text, jyutping, traditional, simplified)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtFTS.Close()

	for _, e := range entries {
		engBytes, err := json.Marshal(e.English)
		if err != nil {
			return err
		}
		ex := e.Examples
		if ex == nil {
			ex = []Example{}
		}
		exBytes, err := json.Marshal(ex)
		if err != nil {
			return err
		}

		if _, err := stmtEntry.ExecContext(
			ctx, e.ID, string(engBytes), e.Jyutping, e.Traditional, e.Simplified,
			e.PartOfSpeech, e.Register, e.Classifier, e.Notes, string(exBytes), e.Source,
			glossKeys(e.English), glossWords(e.English), e.Frequency, joinWordTypes(e.WordTypes),
		); err != nil {
			return err
		}

		if _, err := stmtFTS.ExecContext(
			ctx, e.ID, strings.Join(e.English, " "), e.Jyutping, e.Traditional, e.Simplified,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Search queries entries_fts using FTS5 prefix matching and joins with entries.
func (r *SQLiteRepository) Search(ctx context.Context, q SearchQuery) ([]Entry, error) {
	ftsExpr := buildFTSQuery(q.Query)
	if q.WordType != "" {
		// Filtered lists use whole words, like CountTypes, so the list matches the chip's count.
		ftsExpr = buildFTSWholeWordQuery(q.Query)
	}
	if ftsExpr == "" {
		return []Entry{}, nil
	}

	var args []any
	args = append(args, ftsExpr)

	querySQL := `
		SELECT e.id, e.english_json, e.jyutping, e.traditional, e.simplified,
		       e.part_of_speech, e.register, e.classifier, e.notes, e.examples_json, e.source, e.word_types
		FROM entries_fts f
		JOIN entries e ON e.id = f.id
		WHERE entries_fts MATCH ?
	`
	if q.Register != "" {
		querySQL += ` AND (e.register = ? OR e.register = 'both')`
		args = append(args, q.Register)
	}
	if q.WordType != "" {
		querySQL += ` AND instr(e.word_types, '|' || ? || '|') > 0`
		args = append(args, q.WordType)
	}
	// Ranking tiers:
	//   0 exact: the query equals one meaning (gloss key), the whole Jyutping, or the headword;
	//   1 whole words: every query word appears as a whole word in the glosses or Jyutping;
	//   2 the rest (prefix matches such as "butt" for "but").
	// Within a tier: hand-curated entries first. Among exact matches, how common the word is decides,
	// lowered by how far down the entry's meanings the match is (see senseIndexSQL).
	// Among looser matches relevance decides first, so a common word that merely mentions the
	// query does not jump ahead. Register is the last tie-break.
	// bm25 weights follow the FTS column order: id (unindexed), english_text, jyutping, traditional, simplified.
	headword := strings.TrimSpace(q.Query)
	key := exactKey(q.Query)
	exactSQL := `(instr(e.gloss_keys, '|' || ? || '|') > 0 OR lower(e.jyutping) = ? OR e.traditional = ? OR e.simplified = ?)`
	exactArgs := []any{key, key, headword, headword}
	// 0-based position of the matching meaning in the entry ("if" is meaning #0 of 若果 but
	// about #16 of 要). Each step down costs senseRankPenalty Zipf points, so a word whose main
	// meaning matches beats a very common word where that meaning is buried.
	senseIndexSQL := `(CASE WHEN instr(e.gloss_keys, '|' || ? || '|') > 0 THEN
		length(substr(e.gloss_keys, 1, instr(e.gloss_keys, '|' || ? || '|')))
		- length(replace(substr(e.gloss_keys, 1, instr(e.gloss_keys, '|' || ? || '|')), '|', '')) - 1
		ELSE 0 END)`
	args = append(args, exactArgs...)

	wholeSQL := "0"
	if words := queryWords(q.Query); len(words) > 0 {
		conds := make([]string, len(words))
		for i, w := range words {
			conds[i] = `(instr(e.gloss_words, ?) > 0 OR instr(' ' || lower(e.jyutping) || ' ', ?) > 0)`
			padded := " " + w + " "
			args = append(args, padded, padded)
		}
		wholeSQL = "(" + strings.Join(conds, " AND ") + ")"
	}

	querySQL += `
		ORDER BY
			CASE WHEN ` + exactSQL + ` THEN 0 WHEN ` + wholeSQL + ` THEN 1 ELSE 2 END,
			CASE WHEN e.source = 'custom' THEN 0 ELSE 1 END,
			CASE WHEN ` + exactSQL + ` THEN -(e.freq - ` + senseRankPenalty + ` * ` + senseIndexSQL + `) ELSE 0 END,
			bm25(entries_fts, 0.0, 10.0, 5.0, 2.0, 2.0),
			e.freq DESC,
			CASE WHEN e.register = 'spoken' THEN 0 WHEN e.register = 'both' THEN 1 ELSE 2 END
		LIMIT ?
	`
	args = append(args, exactArgs...)
	args = append(args, key, key, key)
	args = append(args, q.Limit)

	rows, err := r.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEntries(rows)
}

// List returns a paginated slice of dictionary entries and the total matching entry count.
func (r *SQLiteRepository) List(ctx context.Context, q ListQuery) ([]Entry, int, error) {
	var conds []string
	var countArgs []any
	if q.Register != "" {
		conds = append(conds, `(register = ? OR register = 'both')`)
		countArgs = append(countArgs, q.Register)
	}
	if q.WordType != "" {
		conds = append(conds, `instr(word_types, '|' || ? || '|') > 0`)
		countArgs = append(countArgs, q.WordType)
	}
	whereClause := ""
	if len(conds) > 0 {
		whereClause = ` WHERE ` + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM entries`+whereClause, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := `
		SELECT id, english_json, jyutping, traditional, simplified,
		       part_of_speech, register, classifier, notes, examples_json, source, word_types
		FROM entries` + whereClause + `
		ORDER BY
			CASE WHEN source = 'custom' THEN 0 ELSE 1 END,
			CASE WHEN register = 'spoken' THEN 0 WHEN register = 'both' THEN 1 ELSE 2 END,
			id ASC
		LIMIT ? OFFSET ?
	`
	args := append(countArgs, q.Limit, q.Offset)

	rows, err := r.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func scanEntries(rows *sql.Rows) ([]Entry, error) {
	var results []Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, entry)
	}
	return results, rows.Err()
}

// GetByID fetches a single entry by its unique ID.
func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*Entry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, english_json, jyutping, traditional, simplified,
		       part_of_speech, register, classifier, notes, examples_json, source, word_types
		FROM entries
		WHERE id = ?
	`, id)

	entry, err := scanEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entry, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEntry(s rowScanner) (Entry, error) {
	var e Entry
	var engJSON, exJSON, wordTypes string
	err := s.Scan(
		&e.ID, &engJSON, &e.Jyutping, &e.Traditional, &e.Simplified,
		&e.PartOfSpeech, &e.Register, &e.Classifier, &e.Notes, &exJSON, &e.Source, &wordTypes,
	)
	if err != nil {
		return Entry{}, err
	}
	if err := json.Unmarshal([]byte(engJSON), &e.English); err != nil {
		return Entry{}, fmt.Errorf("unmarshal english: %w", err)
	}
	if err := json.Unmarshal([]byte(exJSON), &e.Examples); err != nil {
		return Entry{}, fmt.Errorf("unmarshal examples: %w", err)
	}
	e.WordTypes = splitWordTypes(wordTypes)
	return e, nil
}

// Word types are stored as "|noun|verb|" so one type can be matched with instr().
func joinWordTypes(types []string) string {
	if len(types) == 0 {
		return ""
	}
	return "|" + strings.Join(types, "|") + "|"
}

func splitWordTypes(stored string) []string {
	trimmed := strings.Trim(stored, "|")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "|")
}

// CountTypes counts the query's whole-word matches per word type in one pass. Prefix-only
// matches ("some" for "so") are left out: they would make the counts meaningless.
func (r *SQLiteRepository) CountTypes(ctx context.Context, q SearchQuery) (TypeCounts, error) {
	counts := TypeCounts{ByType: map[string]int{}}
	ftsExpr := buildFTSWholeWordQuery(q.Query)
	if ftsExpr == "" {
		return counts, nil
	}
	cols := make([]string, 0, len(WordTypes)+2)
	for _, t := range WordTypes { // fixed internal list, safe to inline
		cols = append(cols, `COALESCE(SUM(instr(e.word_types, '|`+t+`|') > 0), 0)`)
	}
	cols = append(cols, `COALESCE(SUM(e.word_types = ''), 0)`, `COUNT(*)`)
	querySQL := `SELECT ` + strings.Join(cols, ", ") + `
		FROM entries_fts f JOIN entries e ON e.id = f.id
		WHERE entries_fts MATCH ?`
	args := []any{ftsExpr}
	if q.Register != "" {
		querySQL += ` AND (e.register = ? OR e.register = 'both')`
		args = append(args, q.Register)
	}
	values := make([]int, len(WordTypes)+2)
	dest := make([]any, len(values))
	for i := range values {
		dest[i] = &values[i]
	}
	if err := r.db.QueryRowContext(ctx, querySQL, args...).Scan(dest...); err != nil {
		return TypeCounts{}, err
	}
	for i, t := range WordTypes {
		if values[i] > 0 {
			counts.ByType[t] = values[i]
		}
	}
	counts.Untyped = values[len(WordTypes)]
	counts.Total = values[len(WordTypes)+1]
	return counts, nil
}

// normalizeForExactMatch lowercases and collapses whitespace so "  Thank   You " equals "thank you".
func normalizeForExactMatch(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(raw), " "))
}

// buildFTSQuery turns a search into FTS5 prefix terms: "goi" finds goi1, "wat" finds water.
func buildFTSQuery(raw string) string {
	return buildFTS(raw, false)
}

// buildFTSWholeWordQuery matches whole English/Jyutping words only ("so" does not find "some").
// Chinese tokens stay prefix terms, because the index stores a run of characters as one token.
func buildFTSWholeWordQuery(raw string) string {
	return buildFTS(raw, true)
}

func buildFTS(raw string, wholeASCIIWords bool) string {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '"', '*', ':', '-', '(', ')', '^':
			return ' '
		default:
			return r
		}
	}, strings.TrimSpace(raw))

	tokens := strings.Fields(cleaned)
	if len(tokens) == 0 {
		return ""
	}
	for i, t := range tokens {
		if wholeASCIIWords && isASCII(t) {
			tokens[i] = fmt.Sprintf(`"%s"`, t)
		} else {
			tokens[i] = fmt.Sprintf(`"%s"*`, t)
		}
	}
	return strings.Join(tokens, " ")
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
