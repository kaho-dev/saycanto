# SayCanto — guide for Claude

Open-source **English → Cantonese** dictionary. Jyutping is shown first, every entry carries a
spoken / formal register tag, plus classifiers, usage notes and natural example sentences.
Early stage: the core search → results → word page loop works; data is tiny (10 curated entries).

The engineering rules in AGENTS.md are binding for all agents:

@AGENTS.md

Current state, known bugs and next steps: `docs/STATUS.md` (read it before starting work, update it when you finish).

## Layout
```
backend/            Go 1.23 API + importer (chi, modernc.org/sqlite, FTS5). See backend/CLAUDE.md
  cmd/api           HTTP server; all DI wiring lives in main.go
  cmd/importer      builds dictionary.db from data/ (CC-Canto optional + custom.json overlay)
  internal/<feature>/  model · repository (interface) · sqlite_repo · service · controller
frontend/           Next.js 15 App Router, React 19, Tailwind 3, Vitest. See frontend/CLAUDE.md
  src/app           routes: / (client search page), /word/[id] (server component)
  src/features/dictionary   api · components · hooks · types
  src/shared        api client, UI primitives
data/custom.json    hand-curated entries (CC BY-SA 4.0); the source of truth for quality data
data/CREDITS.md     credits and licences for every data source (shipped with each dictionary release)
Dockerfile          API image for Cloud Run; downloads the released dictionary.db.gz (see docs/DEPLOY.md)
.github/workflows/ci.yml   backend + frontend jobs, both gate on >= 80% coverage
```

## Commands
Run from the folder shown. The shell here is Windows; Git Bash and PowerShell both work.

| Task | Command |
|---|---|
| Build DB | `cd backend && go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db` |
| Build DB + CC-Canto | add `-cccanto ../data/cccanto-webdist.txt` (file not in repo; download from cantonese.org) |
| Run API (:8080) | `cd backend && go run ./cmd/api -db ./dictionary.db -port 8080` (env: `DB_PATH`, `PORT`, `CORS_ORIGIN`) |
| Backend tests | `cd backend && go test -race -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out` |
| Run web (:3000) | `cd frontend && npm run dev` (env: `NEXT_PUBLIC_API_URL`, default `http://localhost:8080`) |
| Frontend tests | `cd frontend && npm run test:coverage` |
| Typecheck / build | `cd frontend && npx tsc --noEmit` / `npm run build` |

Rebuild `dictionary.db` after any change to `data/custom.json` or the schema; the API does not re-import.

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
`source` (`custom` | `cc-canto`) is set by the importer, never written in `custom.json`; curated entries rank higher.
Defined in `backend/internal/dictionary/model.go` and mirrored in `frontend/src/features/dictionary/types.ts`.
**Change both together**, plus `schemaSQL`/`InsertEntries`/`scanEntry` in `sqlite_repo.go` and `CONTRIBUTING.md`.

- `id` convention: Jyutping syllables joined by `-` (`m4-goi1`); importer appends `-2`, `-3` for duplicates.
- `register`: `spoken` (colloquial Cantonese, 口語), `formal` (written/書面語), `both`. Filtering by spoken or formal also returns `both`.
- Jyutping: lowercase, tone number on every syllable (1–6), syllables space-separated.

## Cantonese content rules
- Never invent Jyutping, characters or example sentences you are unsure of. Flag uncertain readings for a native-speaker review instead of guessing; a wrong tone is a wrong word.
- Use Hong Kong traditional characters as `traditional`; make sure `simplified` is the real simplified form, not a copy.
- Colloquial characters (唔, 嘅, 咗, 喺, 冇, 嘢) belong to `spoken` entries; note the formal equivalent in `notes` where useful.
- Data contributions are CC BY-SA 3.0; do not paste entries from sources with incompatible licences (e.g. words.hk content, commercial dictionaries).

## Working rules
- Keep the >= 80% coverage gate green on both sides; add tests with every feature (mocks for service/controller, `:memory:` SQLite for repos).
- No global DB handles; new dependencies go through `New...` constructors wired in `cmd/api/main.go`.
- Frontend data fetching only via `src/shared/api/client.ts` → feature `api/` module → hooks.
- Generated/ignored: `backend/dictionary.db`, `coverage*`, `frontend/.next`, `node_modules`. Don't commit them.
- The folder is not a git repo yet. Ask before `git init`, adding a remote or pushing.
