# SayCanto — guide for Claude

Open-source **English → Cantonese** dictionary. Jyutping is shown first, with spoken / formal register tags,
word types, classifiers, usage notes and natural example sentences.
**Live:** https://saycanto.vercel.app (Vercel) + API on Google Cloud Run. ~123k entries (CC-Canto + CC-CEDICT with
Cantonese readings + 24 curated). Repo: https://github.com/kaho-dev/saycanto (branch `main`).

The engineering rules in AGENTS.md are binding for all agents:

@AGENTS.md

**Start every session by reading `docs/HANDOVER.md`** (latest session, decisions, traps, next steps), then
`docs/STATUS.md` (current snapshot). At the end of a session add a dated section to HANDOVER.md and refresh STATUS.md.
Deployment and accounts: `docs/DEPLOY.md`.

## Layout
```
backend/            Go 1.23 API + importer (chi, modernc.org/sqlite, FTS5). See backend/CLAUDE.md
  cmd/api           HTTP server; all DI wiring lives in main.go
  cmd/importer      builds dictionary.db from data/ (sources + frequencies + word types + custom.json overlay)
  internal/<feature>/  model · repository (interface) · sqlite_repo · service · controller
frontend/           Next.js 15 App Router, React 19, Tailwind 3, Vitest. See frontend/CLAUDE.md
  src/app           routes: / (client search page), /word/[id] (server component)
  src/features/dictionary   api · components · hooks · types
  src/shared        api client, UI primitives
data/custom.json    hand-curated entries (CC BY-SA 4.0); the source of truth for quality data
data/README.md      every data source, download link and the full importer command (source files are gitignored)
data/CREDITS.md     credits and licences for every data source (shipped with each dictionary release)
Dockerfile          API image for Cloud Run; downloads the released dictionary.db.gz, SHA-256 checked
scripts/check_words.py   audits a word list (scripts/wordlists/*.tsv) against the running API
docs/               HANDOVER.md, STATUS.md, DEPLOY.md, conjunction-check.md
.github/workflows/ci.yml   backend + frontend (>= 80% coverage) + docker (build and smoke-test the image)
```

## Commands
Run from the folder shown. The shell here is Windows; Git Bash and PowerShell both work.

| Task | Command |
|---|---|
| Build curated-only DB | `cd backend && go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db` |
| Build full DB (~15 s) | `cd backend && go run ./cmd/importer -cccanto ../data/cccanto-webdist.txt -cedict ../data/cedict_1_0_ts_utf-8_mdbg.zip -cedict-readings ../data/cccedict-canto-readings-150923.txt -wordfreq ../data/large_zh.msgpack.gz -hkcancor ../data/hkcancor-utf8.zip -wordnet ../data/english-wordnet-2025.zip -custom ../data/custom.json -out ./dictionary.db` |
| Get the released DB instead | `gh release download data-2026-11-27 -R kaho-dev/saycanto -p dictionary.db.gz`, then gunzip to `backend/dictionary.db` |
| Run API (:8080) | `cd backend && go run ./cmd/api -db ./dictionary.db -port 8080` (env: `DB_PATH`, `PORT`, `CORS_ORIGIN`) |
| Backend tests | `cd backend && go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out` (CI adds `-race`, which needs CGO, so skip it on Windows) |
| Run web (:3000) | `cd frontend && npm run dev` (env: `NEXT_PUBLIC_API_URL`, default `http://localhost:8080`) |
| Frontend tests | `cd frontend && npm run test:coverage` |
| Typecheck / build | `cd frontend && npx tsc --noEmit` / `npm run build` |

| Word-list audit | `python scripts/check_words.py scripts/wordlists/conjunctions.tsv --out docs/conjunction-check.md` (API on :8080) |

Rebuild `dictionary.db` after any change to `data/custom.json` or the schema; the API does not re-import, and it holds
the file open, so stop it first. Shipping new data to production = a new GitHub release + `DICT_URL`/`DICT_SHA256` in
the `Dockerfile` (steps in `docs/DEPLOY.md`).

## API (all GET, JSON)
| Route | Params | Returns |
|---|---|---|
| `/api/health` | – | `{"status":"ok"}` |
| `/api/search` | `q`, `register` (spoken\|formal\|both), `type` (word type), `limit` (default 20, max 50) | `{query, total, entries[], typeCounts{type: n}, untyped}` |
| `/api/words` | `register`, `type`, `limit`, `offset` | `{total, limit, offset, entries[]}` |
| `/api/words/{id}` | – | one `Entry`; 404 `{"error":...}` if missing |

Errors are always `{"error": "..."}` via `internal/shared/httputil`. Invalid register or type → 400.
`total` and `typeCounts` count whole-word matches ("so" does not count "some"); `typeCounts` ignores the `type` filter.
Word types: noun, verb, adjective, adverb, pronoun, preposition, conjunction, classifier, particle, interjection,
number, expression (`dictionary.WordTypes`, mirrored in `frontend/src/features/dictionary/wordTypes.ts`).

## Data model (one shape everywhere)
`Entry` = `id, english[], jyutping, traditional, simplified, partOfSpeech?, register?, classifier?, notes?, examples[]{jyutping, cantonese, english}, source?, wordTypes?[]`.
`source` (`custom` | `cc-canto` | `cc-cedict`) and `wordTypes` are set by the importer, never written in `custom.json`;
curated entries rank higher. `Frequency` exists in Go only (not in the API).
Defined in `backend/internal/dictionary/model.go` and mirrored in `frontend/src/features/dictionary/types.ts`.
**Change both together**, plus `schemaSQL`/`InsertEntries`/`scanEntry` in `sqlite_repo.go` and `CONTRIBUTING.md`.

- `id` convention: Jyutping syllables joined by `-` (`m4-goi1`); importer appends `-2`, `-3` for duplicates.
- `register`: `spoken` (colloquial Cantonese, 口語), `formal` (written/書面語), `both`. Filtering by spoken or formal also returns `both`.
- Jyutping: lowercase, tone number on every syllable (1–6), syllables space-separated.

## Cantonese content rules
- Never invent Jyutping, characters or example sentences you are unsure of. Flag uncertain readings for a native-speaker review instead of guessing; a wrong tone is a wrong word.
- Use Hong Kong traditional characters as `traditional`; make sure `simplified` is the real simplified form, not a copy.
- Colloquial characters (唔, 嘅, 咗, 喺, 冇, 嘢) belong to `spoken` entries; note the formal equivalent in `notes` where useful.
- Data contributions are CC BY-SA 4.0; do not paste entries from sources with incompatible licences (e.g. words.hk content, commercial dictionaries).
- New curated entries go on the native-speaker review list in `docs/STATUS.md` until reviewed.

## Working rules
- Keep the >= 80% coverage gate green on both sides; add tests with every feature (mocks for service/controller, `:memory:` SQLite for repos).
- No global DB handles; new dependencies go through `New...` constructors wired in `cmd/api/main.go`.
- Frontend data fetching only via `src/shared/api/client.ts` → feature `api/` module → hooks.
- Generated/ignored: `backend/dictionary.db`, `coverage*`, `frontend/.next`, `node_modules`, `dist/`, and the
  downloaded data files in `data/`. Don't commit them.
- Pushing to `main` deploys to production (Cloud Run + Vercel). Commit or push only when the owner asks.
- Downloads, publishing (releases, posts), account sign-ups, sign-ins, verification codes and anything costing money
  need the owner's explicit OK; sign-ins and codes are theirs to do. Never script the CC-CEDICT download (MDBG forbids it).
- The owner's Google Cloud and Vercel sessions are in their Chrome (Claude in Chrome), not the built-in browser.
