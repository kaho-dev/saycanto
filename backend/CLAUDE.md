# Backend notes (Go)

Module `github.com/kaho-dev/saycanto/backend`, Go 1.23 (CI pins 1.23; newer local toolchains are fine).
Pure-Go SQLite (`modernc.org/sqlite`, driver name `"sqlite"`), so `CGO_ENABLED=0` builds work.

## Request path
`cmd/api/main.go` → chi router + middleware (RequestID, RealIP, Logger, Recoverer, 10 s Timeout, CORS)
→ `dictionary.Controller` (parse query, map errors to status) → `dictionary.Service` (trim, validate register,
clamp limit 1..50, nil → empty slice) → `dictionary.Repository` (`SQLiteRepository`).

## Storage
- `entries`: one row per entry; `english` and `examples` are stored as JSON text columns.
- `entries_fts` (FTS5, `unicode61`): columns `id UNINDEXED, english_text, jyutping, traditional, simplified`.
- `InsertEntries` **wipes both tables** and rewrites everything in one transaction; the importer is a full rebuild, not an upsert.
- Search: `buildFTSQuery` strips FTS syntax chars and turns each token into a prefix term (`"goi"*`), ANDed.
  Results are ordered by: exact-match tier (gloss = query or "to " + query, whole Jyutping, whole headword)
  → curated (`source = 'custom'`) → for exact matches, commonness (`freq`, whole Zipf points) → `bm25` → `freq`
  → register (spoken → both → formal). Tiers: exact meaning (`gloss_keys`) / whole Jyutping / headword → whole words
  (`gloss_words`) → prefix. See `internal/dictionary/gloss.go`.
- Word types (`entries.word_types`, stored "|noun|verb|") come from `internal/importer/wordtypes.go`: curated
  partOfSpeech, leading labels like "(noun)", "(idiom)" anywhere, HKCanCor POS tags (≥20% of a word's uses), and
  "to …" → verb, "classifier for …" → classifier, "… particle" → particle. Then, for still-untyped entries only,
  Open English WordNet (`internal/importer/wordnet.go`): first 3 meanings, ≤3 words, verb senses ignored for bare
  meanings, type given only if it has ≥70% of the word's noun/adjective/adverb senses.
  Search with a `type` filter and `CountTypes` use whole-word FTS (`buildFTSWholeWordQuery`); search and count run
  concurrently in the service.
- Importer flags: `-cccanto`, `-cedict` + `-cedict-readings`, `-wordfreq`, `-hkcancor`, `-wordnet`, `-custom`, `-out`
  (full command in `data/README.md`). Readings and frequencies are never guessed.
- Schema changes: `InitSchema` runs `CREATE ... IF NOT EXISTS` and then `addColumnIfMissing` for columns added later
  (`source`, `gloss_keys`, `gloss_words`, `freq`, `word_types`), so older `dictionary.db` files upgrade in place (derived
  columns are only filled by a re-import). Add new columns the same way. When changing ranking, run
  `TestSQLiteRepositorySearchRanking`, `TestSearchRankingTiers`, `TestSearchPrefersCommonWords`, `TestSearchPrefersMainMeaning`.
- `bm25()` weights map to FTS columns **in declaration order, including the UNINDEXED `id`** (currently `0, 10, 5, 2, 2`).

## Importer (`cmd/importer/main.go`, `internal/importer/`)
1. `ParseCCCanto` (trailing `# notes` and empty `[]` pinyin handled; every entry tagged `spoken`).
2. `ParseCantoReadings` + `ParseCEDICT` (`cedict.go`; zip read directly; v1 and v2 pinyin; skips headwords with no
   Cantonese reading and cross-reference-only senses) → `CombineBase` (same characters + Jyutping → meanings merged).
3. `ParseCustomJSON` → `ValidateOverlay` (unique, non-empty ids) → `MergeEntries` (a curated entry replaces every base
   entry with the same `traditional|jyutping` key; curated ids never change; colliding base ids get the next free `-N`).
4. `LoadWordfreq` (minimal msgpack decoder) + `LoadHKCanCor` → `ApplyFrequencies` (`frequency.go`).
5. `InferWordTypes` (`wordtypes.go`) → `InferWordTypesFromWordNet` (`wordnet.go`, untyped entries only).
6. `BuildDatabase` (full rebuild). The importer logs counts for every step; check them after a rebuild.
On Windows run tests without `-race` (needs CGO); CI keeps it.

## Tests
- `service_test.go` / `controller_test.go` use hand-written mocks of the interfaces.
- `sqlite_repo_test.go` and `importer_test.go` use `sql.Open("sqlite", ":memory:")`.
- Every new exported behaviour needs a test; CI fails under 80 % total for `./internal/...`.

## Adding a feature
Create `internal/<feature>/` with `model.go`, `repository.go` (interface), `sqlite_repo.go`, `service.go`,
`controller.go` (+ tests), expose `RegisterRoutes(chi.Router)`, wire it in `cmd/api/main.go`.
