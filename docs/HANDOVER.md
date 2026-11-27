# Handover notes

Read this first when starting a new session, then `docs/STATUS.md` (current state) and the CLAUDE.md files.
Add a new dated section at the top at the end of each working session.

---

## Session 2026-11-27 — from Antigravity scaffold to live site

### Where things stand
- **Live:** website https://saycanto.vercel.app · API https://saycanto-api-417998879746.europe-west1.run.app
- **Repo:** https://github.com/kaho-dev/saycanto, branch `main`, last commit `cd29ece`, working tree clean.
  CI (backend, frontend, docker) green on every push.
- **Dictionary:** 122,998 entries (CC-Canto + CC-CEDICT with Cantonese readings + 24 curated), with frequency scores
  and word types (61% typed). Published as GitHub release `data-2026-11-27` (`dictionary.db.gz`, SHA-256 in `Dockerfile`).
- **Local servers:** stopped. Restart commands in CLAUDE.md.

### What was done (in order)
1. Read the Antigravity scaffold ("JyutDict"), wrote the CLAUDE.md files and `docs/STATUS.md`.
2. Fixed the importer crash on duplicate CC-Canto readings (ID collisions) and two parser gaps (trailing `# comments`,
   empty `[]` pinyin) so all of CC-Canto imports.
3. Search ranking rebuilt: exact meaning → whole word → prefix tiers; curated first; word frequency (wordfreq + HKCanCor);
   main-meaning penalty; label/numbering clean-up of meanings.
4. Added CC-CEDICT + Pleco's Cantonese readings (everyday words such as 通常, 雖然, 如果 were missing before).
5. 🔊 Listen button (browser Cantonese voice; hidden when the device has none).
6. Conjunction audit (`scripts/check_words.py`, 55 words): 45/55 in the top 5; 14 curated conjunctions added.
7. Word types + filter chips with counts (HKCanCor tags, labels, "to …" verbs, Open English WordNet for nouns/adjectives).
8. Renamed **JyutDict → SayCanto** (name taken by jyutdict.org and close to Jyut Dictionary); Go module
   `github.com/kaho-dev/saycanto/backend`.
9. Git repo created and pushed; MIT `LICENSE`; data credits (`data/CREDITS.md`, site footer).
10. Deployed: Google Cloud Run (API) + Vercel (website); API locked to the site with `CORS_ORIGIN`; image clean-up rules.

### Decisions the owner made (keep to them)
- Name **SayCanto**; repo `kaho-dev/saycanto`; code MIT, data **CC BY-SA 4.0** (CC-Canto's 3.0 allows the upgrade).
- Hosting: **Cloud Run** (card on file, fast wake-up) for the API, **Vercel Hobby** for the site. Free-tier usage only.
- Every data download was confirmed by the owner first (file, source, size). Keep asking before new downloads.
- CC-CEDICT must be downloaded **by hand** (MDBG forbids scripted access) — never script it.
- Readings are never guessed; uncertain Cantonese goes to a native-speaker review list.

### Accounts and services (owner's; sign-ins and codes are always the owner's to do)
| Service | What | Notes |
|---|---|---|
| GitHub `kaho-dev` | repo, releases, CI | `gh` CLI is signed in on this PC |
| Google Cloud | project `saycanto-516012`, service `saycanto-api`, `europe-west1` | billing account "My Billing Account 2", €1/month budget alert; Cloud Build trigger deploys on push to `main` |
| Vercel | team "MarshalBuilder" (Hobby), project `saycanto`, root `frontend` | env `NEXT_PUBLIC_API_URL` = the Cloud Run URL; deploys on push to `main` |

The owner's Google Cloud and Vercel sessions are in their **Chrome** (use Claude in Chrome), not the built-in browser.

### Traps found this session
- **Windows:** `go test -race` needs CGO, so run without `-race` locally (CI runs it). Python here prints Chinese only
  after `sys.stdout.reconfigure(encoding="utf-8")` (`-I` ignores `PYTHONIOENCODING`).
- **Writing regexes through Python heredocs:** `\b` became a backspace once. Check with `cat -A` or use the Edit tool.
- **The API holds `dictionary.db` open:** stop it before re-importing.
- **Docker Desktop** is not running on this PC; the CI `docker` job is the image test.
- **`CORS_ORIGIN` is a single origin.** Vercel preview deployments can't call the API from the browser. A custom domain
  will need a small code change to accept a list (see next steps).
- **Cloud Run console:** "Service account" is required on create (default compute SA was chosen); settings are edited on
  the service's Containers tab → "View diff & redeploy"; env vars survive automatic Cloud Build deploys.
- **Curated entries replace source entries** with the same characters + Jyutping, including their other senses
  (curated 水 "water" dropped CC-Canto's slang 水 "money"). Put extra senses in `custom.json`.

### Waiting on the owner
- **Native-speaker review:** the 14 curated conjunctions added this session (`data/custom.json`), CC-Canto's 為咗
  **wai4** zo2 (probably wai6), 儘管 zeon2 vs zeon6. See `docs/conjunction-check.md`.
- **Custom domain:** saycanto.com / .app / .org / .io / .dev had no DNS on 27 Nov (likely unregistered). Owner plans
  to buy one; then connect it (steps below).

### Suggested next steps
1. **Custom domain** (after purchase): make `CORS_ORIGIN` accept a comma-separated list (`httputil.CORSMiddleware`,
   echo the matching `Origin`, add `Vary: Origin`, tests), add the domain in Vercel, set DNS, update `CORS_ORIGIN`,
   README, footer, `docs/DEPLOY.md`.
2. **Search speed for 1–2 letter queries** ("a" ≈ 0.6 s): minimum length for prefix terms or for counts.
3. **Chinese-character search inside words** (該 → 唔該): FTS5 trigram tokenizer or a LIKE fallback.
4. **Compact Jyutping** (`mgoi`, `m4goi1`): syllable splitting before search.
5. **Group duplicate readings** of one character (吃 jaak3 / hek3) and show all examples on the word page.
6. **More curated words** (the ~200 most-searched English words), reusing `scripts/check_words.py` with new word lists.
7. **API fails fast on an empty/missing database** (STATUS bug 5).
8. Update `actions/checkout@v4` / `setup-node@v4` / `setup-go@v5` before GitHub drops Node 20 runners.

### How to resume
```powershell
cd "C:\Users\Kaho\Documents\Open Source Projects\saycanto"
```
- Local full dictionary needs the six data files in `data/` (gitignored; listed in `data/README.md`). They are on this PC.
- Quick start without the data files: download the release DB instead —
  `gh release download data-2026-11-27 -R kaho-dev/saycanto -p dictionary.db.gz` then gunzip into `backend/dictionary.db`.
- Then run the API and web commands from CLAUDE.md (or the `api` / `web` entries in `.claude/launch.json`).
