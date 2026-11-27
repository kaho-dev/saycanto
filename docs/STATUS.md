# SayCanto status

_Last reviewed: 27 Nov 2026 (first read-through by Claude; project originally scaffolded with Antigravity)._

## Works today (checked)
- Backend tests pass: total coverage 92.1 % (dictionary 88.9, health 100, importer 95.8, httputil 100). `go vet` clean.
  Locally `-race` needs CGO (a C compiler), so run without `-race` on Windows; CI on Linux runs with it.
- Frontend: 21 Vitest tests in 6 files pass; coverage 99 % statements / 94 % branches on measured files; `tsc --noEmit` clean.
- Search by English or Jyutping prefix, register filter, starter-word list, word detail page, copy-Jyutping button.
- `data/custom.json`: 10 curated entries.
- Full CC-Canto import (27 Nov): `data/cccanto-webdist.txt` (CC-Canto 2017-02-02, from https://cantonese.org/cccanto-170202.zip,
  gitignored) → 34,334 parsed + 10 curated = **34,335 entries**, `backend/dictionary.db` ≈ 11.6 MB. Spot-checked searches:
  thank you, water, m4 goi1, sik6 faan6, 嘢, where.

Not checked yet: `npm run build`, `npm run lint`, the Docker image, an end-to-end run of API + web together, a real CC-Canto import.

- 🔊 Listen button (27 Nov): `src/shared/ui/ListenButton.tsx` + `src/shared/speech/cantoneseVoice.ts`. Uses the browser's
  built-in Cantonese voice (zh-HK / yue tags), reads the **characters** of the headword and first example, rate 0.85.
  Hidden when the device has no Cantonese voice (never falls back to Mandarin). Caveat: for characters with several
  readings the voice may not match the Jyutping shown. Next: optional `audio` field with native recordings for curated entries.
- Starter words now list curated entries first (were odd CC-Canto rows sorted by id, e.g. "1 mok1 pei4 ngau4").

- **Full dictionary (27 Nov)**: CC-Canto + CC-CEDICT (2026-11-27 release, downloaded by hand; MDBG forbids scripted access)
  + Pleco's CC-CEDICT Cantonese readings (2015) + 10 curated = **122,996 entries**, `dictionary.db` ≈ 60 MB, import ≈ 13 s.
  CC-CEDICT: 125,479 lines → 95,538 entries; 25,425 skipped (no Cantonese reading; never guessed), 4,516 cross-reference-only;
  6,874 merged into CC-Canto entries with the same characters + Jyutping. CC-CEDICT entries have no register tag (`source = cc-cedict`).
  Data licence is now CC BY-SA 4.0 (README, CONTRIBUTING, data/README updated).
- **Search tiers (27 Nov)**: exact meaning (labels like "(conjunction)" ignored, lists split on ; , /) → whole word → prefix.
  Stored per entry as `gloss_keys` / `gloss_words` (see `internal/dictionary/gloss.go`). "but" → 但係/但是, "if" → 假如,
  "although" → 雖說/雖然, "school" → 學校, "computer" → 電腦.

- **Conjunction check (27 Nov)**: `scripts/check_words.py` + `scripts/wordlists/conjunctions.tsv` (55 words) → report in
  `docs/conjunction-check.md`. After fixes: 45/55 in the top 5 of their English search, none missing. Fixes: sense
  numbering ("1. or") split into meanings; exact matches lowered 0.3 Zipf per position down the meaning list; 14 curated
  conjunctions added to `custom.json` (now 24 curated entries).
- **Needs native-speaker review**: the 14 new curated conjunctions (meanings, registers, examples); CC-Canto 為咗 wai4 zo2
  (probably wai6); 儘管 zeon2 vs zeon6.

- **Word-type filter (27 Nov)**: 12 types inferred at import (see backend/CLAUDE.md). With Open English WordNet filling
  untyped entries: **75,407 of 122,998 entries (61%)**: noun 34,225 · verb 27,503 · adjective 8,484 · expression 5,082 ·
  adverb 1,736 · classifier 436 · particle 125 · number 108 · pronoun 100 · conjunction 83 · preposition 49 ·
  interjection 31. Accuracy check of WordNet-only types: 40 random entries, 39 right; the 1 error (打簿 "verb; …" → noun)
  led to a fix (bare label words count as labels). Some types come straight from odd source labels
  (CC-Canto: 去 "(adjective) passed away", 唔 "(verb) don't").
  UI: chips with counts under the search box; filtered lists use whole-word matching so they agree with the counts.
  Search speed with counts: most 20–40 ms; "so" ≈ 0.18 s, "and" ≈ 0.2 s, "a" ≈ 0.6 s.

- **Live (27 Nov 2026)**: website https://saycanto.vercel.app (Vercel), API on Cloud Run (europe-west1), both
  redeploy on every push to `main`. API restricted to the website via `CORS_ORIGIN`. Details in `docs/DEPLOY.md`.

## Known bugs
1. ~~**Importer crashes with real CC-Canto data.**~~ **Fixed 27 Nov.** A curated entry now replaces *every* base entry
   with the same `traditional|jyutping` key, curated ids are kept as written, colliding base ids are renumbered
   (`-N`), and the importer refuses a `custom.json` with duplicate or empty ids (`ValidateOverlay`).
   Regression tests in `internal/importer/importer_test.go`.
2. ~~**bm25 weights shifted by one.**~~ **Fixed 27 Nov**: now `bm25(entries_fts, 0, 10, 5, 2, 2)` (id, english, jyutping, trad, simp).
3. ~~`total` in `/api/search` is the page size~~ **Fixed 27 Nov**: `total` is now the number of whole-word matches.
4. ~~**Register outranks relevance.**~~ **Fixed 27 Nov**: search orders by exact match first (a gloss equal to the query or
   "to <query>", the whole Jyutping, or the whole headword), then bm25, then register as a tie-break. "water" → 水 first,
   "eat" → 食嘢, "where" → 邊度. Tests in `sqlite_repo_test.go` (`TestSQLiteRepositorySearchRanking`).
   **Curated boost added 27 Nov**: entries now carry `source` (`custom` | `cc-canto`); among equally exact matches curated
   ones come first ("eat" → 食飯, "sorry" → 對唔住, "bye" → 再見). An exact CC-Canto match still beats a curated partial one.
   Still open: one-letter queries like "a" take
   ~175 ms on 34k rows (others 1–15 ms), so consider a 2-character minimum.
5. **API with a missing DB path silently serves an empty dictionary** (`sql.Open` creates the file, `InitSchema` makes empty tables).
   Should fail fast or warn when `entries` is empty.

## Limitations to decide on
- ~~No word-frequency signal~~ **Added 27 Nov**: each entry gets a Zipf frequency (`entries.freq`) = max(wordfreq zh by
  simplified form, HKCanCor by characters + Jyutping); 83,321 of 122,996 entries scored. Exact matches are ordered by
  whole Zipf points, looser matches by relevance first. A traditional = simplified headword that shares its simplified
  form with another word (于 vs 於) gets −1.5. Results: usually → 通常, because → 因為, although → 雖然, if → 如果,
  or → 或者, but → 但係, thing → 嘢, go → 去. The "fewer meanings first" tie-break was removed.
  Known quirks: meanings split on commas, so "earth, water, fire, wind" makes 四大 an exact match for "water";
  single-character literary senses (率, 蓋, 乎) still appear in top 5 for some words; words absent from both lists rank as rare.
- Speed on 123k entries: most searches 2–20 ms; very short prefixes are slow ("a" ≈ 0.5 s, "the" ≈ 0.2 s, "or" ≈ 0.15 s).

- Chinese-character search: `unicode61` treats a run of Han characters as one token, so 該 alone will not find 唔該
  (prefix 唔 will). Consider the FTS5 `trigram` tokenizer or a LIKE fallback for CJK queries. (Reasoned from SQLite docs; not tested.)
- Jyutping search needs spaces (`m4 goi1`, `m goi` work; `mgoi` / `m4goi1` don't). Tone-less matching works by prefix only.
- CC-Canto entries all get `register: "spoken"`, no examples, no part of speech.
- ~~CC-Canto lines with trailing comments were dropped~~ **Fixed 27 Nov**: 8,882 lines end in `# adapted from cc-cedict` and
  60 have an empty Mandarin reading `[]`; both now parse. 1 source line is malformed upstream (陳慧琳, stray `}`) and is skipped.
- Duplicate readings of one character (佝 as gau1/kau3/keoi1, 吃 as jaak3/hek3) appear as separate cards; consider grouping.
- A curated entry replaces CC-Canto rows with the same characters + Jyutping, including their other senses. Example:
  curated 水 seoi2 "water" dropped CC-Canto's 水 = (slang) money. Add such senses to `custom.json` when curating.
- EntryCard only shows the first example sentence; the word page reuses the same card (no extra detail yet).
- No audio / pronunciation, no tone colouring, no "spoken ↔ formal" pairs linking.

## Housekeeping
- Not a git repository yet. No `LICENSE` file although README says MIT (code) + CC BY-SA 3.0 (data).
- Repository: https://github.com/kaho-dev/saycanto (Go module `github.com/kaho-dev/saycanto/backend`).
- **Renamed JyutDict → SayCanto (27 Nov 2026)** because "Jyutdict" is used by jyutdict.org (泛粵大典, active, open source) and is
  close to Jyut Dictionary (github.com/aaronhktan/jyut-dict, jyutdictionary.com). Checked free on 27 Nov: GitHub repos/account,
  npm, PyPI; no DNS for saycanto.com/.org/.app/.dev/.io. Not checked: trademarks, social handles (X, Instagram, TikTok).
- Deployment: API on Google Cloud Run (root `Dockerfile` downloads the released `dictionary.db.gz`, SHA-256 checked;
  `--build-arg DICT_URL=` builds curated-only), website on Vercel. Steps in `docs/DEPLOY.md`. Data credits in
  `data/CREDITS.md` and the site footer. CI `docker` job builds and smoke-tests the image.
- Build artefacts on disk (`.next/`, `coverage/`, `coverage.out`, `dictionary.db`) are gitignored.

## Suggested next steps (in order)
1. `git init`, add `LICENSE` (MIT) and `data/LICENSE` (CC BY-SA 3.0), first commit.
2. Fix bugs 1–3 with regression tests.
3. ~~Real CC-Canto import~~ (done); decide register default and ranking (bug 4), add an exact-match boost.
4. CJK + compact-Jyutping search.
5. Grow `custom.json` (target the most common ~200 English lookups), with a validation step in the importer
   (tone numbers present, required fields, unique ids).
6. Richer word page (all examples, related words), then deployment.
