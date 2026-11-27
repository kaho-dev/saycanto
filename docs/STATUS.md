# SayCanto status

_Current snapshot, last updated 27 Nov 2026. History and session notes: `docs/HANDOVER.md`._

## Live
- Website: https://saycanto.vercel.app (Vercel). API: https://saycanto-api-417998879746.europe-west1.run.app (Cloud Run).
- Both redeploy on every push to `main`. API only accepts the website's origin (`CORS_ORIGIN`). Setup: `docs/DEPLOY.md`.

## Health (checked 27 Nov)
- Backend: all tests pass, ~95% coverage (`./internal/...`), `go vet` clean. Run without `-race` on Windows (needs CGO).
- Frontend: 39 Vitest tests pass, `tsc --noEmit` clean, `npm run build` passes (CI and Vercel).
- CI (`.github/workflows/ci.yml`): backend, frontend and docker (builds the image with the released dictionary and
  smoke-tests a search). All green on the latest push.

## Data
| | |
|---|---|
| Entries | **122,998** = CC-Canto (34,334) + CC-CEDICT with Cantonese readings (88,664 new, 6,874 merged) + 24 curated |
| Frequency | 83,323 entries scored (Zipf, max of wordfreq zh and HKCanCor) |
| Word types | 75,407 entries (61%): noun 34,225 · verb 27,503 · adjective 8,484 · expression 5,082 · adverb 1,736 · classifier 436 · particle 125 · number 108 · pronoun 100 · conjunction 83 · preposition 49 · interjection 31 |
| Release | GitHub release `data-2026-11-27`: `dictionary.db.gz` (41.7 MB, ~89 MB unpacked) + `CREDITS.md` |
| Licence | Data CC BY-SA 4.0 (`data/CREDITS.md`); code MIT (`LICENSE`) |

Source files (all gitignored, on the owner's PC, list in `data/README.md`): CC-Canto 2017-02-02, CC-CEDICT 2026-11-27
(hand-downloaded), CC-CEDICT Cantonese readings 2015-09-23, wordfreq `large_zh`, HKCanCor, Open English WordNet 2025.
CC-CEDICT words with no Cantonese reading (25,425) are skipped, never guessed.

## Features
- Search English, Jyutping (with or without tones, space-separated) or characters (prefix). Ranking: exact meaning →
  whole word → prefix; curated first; among exact matches commonness minus 0.3 per position down the meaning list.
- Filters: register (spoken / formal; `both` matches either) and word type (chips with counts, whole-word matches).
- Starter words (curated first), word pages (`/word/[id]`, server-rendered), copy Jyutping, 🔊 Listen (browser
  Cantonese voice; hidden when the device has none; reads characters, so multi-reading characters may not match the Jyutping).
- Credits footer; `scripts/check_words.py` audits a word list against the live search (`docs/conjunction-check.md`).

## Open bugs
1. **Missing/empty database is served silently**: `sql.Open` creates the file and `InitSchema` makes empty tables.
   The API should fail fast or warn when `entries` is empty.
2. **Single CORS origin**: preview deployments and a future custom domain can't call the API until `CORS_ORIGIN` accepts a list.

## Limitations
- Speed: most searches 2–40 ms; 1–2 letter queries are slow ("a" ≈ 0.6 s, "and" ≈ 0.2 s) because prefix matching scans
  many rows and the type counts run over them.
- Chinese search only matches from the start of a word (該 doesn't find 唔該): `unicode61` keeps a run of Han characters
  as one token. Fix: FTS5 trigram tokenizer or a LIKE fallback.
- Jyutping needs spaces (`m goi` works; `mgoi`, `m4goi1` don't).
- Meanings are split on commas, so lists like "earth, water, fire, wind" make 四大 an exact match for "water".
- Some single-character literary senses (率, 蓋, 乎) still reach the top 5; words absent from both frequency lists rank as rare.
- CC-Canto entries are all tagged `spoken`; CC-CEDICT entries have no register. Some word types come from odd source
  labels (CC-Canto 去 "(adjective) passed away", 唔 "(verb) don't").
- Duplicate readings of one character appear as separate cards (吃 jaak3 / hek3). EntryCard shows only the first example.
- A curated entry replaces source entries with the same characters + Jyutping, including their other senses
  (curated 水 "water" dropped CC-Canto's slang "money"). Put extra senses into `custom.json`.

## Needs native-speaker review
- The 14 curated conjunctions added 27 Nov in `data/custom.json` (要係, 如果, 同, 所以, 因此, 噉, 仲有, 跟住, 不如, 唔係,
  好似, 同時, 一係, 為咗): meanings, registers, example sentences.
- CC-Canto 為咗 **wai4** zo2 (probably wai6; the curated entry uses wai6). 儘管 zeon2 vs zeon6.

## Next steps
See "Suggested next steps" in `docs/HANDOVER.md` (custom domain + CORS list first).
