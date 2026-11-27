# Contributing to SayCanto

Thanks for helping build a better English → Cantonese dictionary! Anyone can open a pull request (PR), but
**nothing reaches `main` without passing CI and a review by a maintainer** (see [How PRs are reviewed](#how-prs-are-reviewed)).
Merging to `main` deploys straight to the live site, so we keep this strict.

## Before you start
- **Open an issue first** for anything bigger than a typo or a few dictionary entries, so we can agree on the
  approach before you spend time on it. PRs for features nobody discussed may be closed.
- Keep each PR to **one change**. Small PRs get reviewed quickly; large mixed ones get sent back.
- Fork the repo and work on a branch in your fork. Never commit to `main`.

## Adding or fixing dictionary entries
1. Edit `data/custom.json`. Each entry looks like this:
   ```json
   {
     "id": "m4-goi1",
     "english": ["thank you", "please", "excuse me"],
     "jyutping": "m4 goi1",
     "traditional": "唔該",
     "simplified": "唔该",
     "partOfSpeech": "expression",
     "register": "spoken",
     "notes": "Used for thanking someone for a service or favour …",
     "examples": [
       { "jyutping": "m4 goi1, maai4 daan1.", "cantonese": "唔該，埋單。", "english": "Excuse me, check please." }
     ]
   }
   ```
   - `id`: the Jyutping syllables joined by `-` (add `-2`, `-3` for another sense of the same word).
   - `jyutping`: lowercase, a tone number (1–6) on every syllable, syllables separated by spaces.
   - `traditional`: Hong Kong characters. `simplified`: the real simplified form (not a copy).
   - `register`: `spoken` (口語), `formal` (書面語) or `both`.
   - Optional: `partOfSpeech`, `classifier`, `notes`, `examples`. Don't add `source` or `wordTypes`; the importer sets them.
2. Check it:
   ```bash
   python scripts/validate_custom.py
   cd backend && go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db
   ```
3. Open a PR and fill in the template, including where the data came from.

**Data rules.** Never guess a reading: a wrong tone is a wrong word. If you're not a native or fluent speaker,
say so in the PR and we'll ask a native speaker to check it. Only use your own knowledge or CC BY-SA–compatible
sources; **don't copy from words.hk, Pleco's paid dictionaries or other commercial dictionaries.**

## Code contributions
1. Read [AGENTS.md](AGENTS.md) (architecture, dependency injection, testing rules).
2. Run the same checks as CI before you push:
   ```bash
   # Backend
   cd backend
   gofmt -l .            # must print nothing
   go vet ./...
   go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out   # total >= 80%

   # Frontend
   cd frontend
   npx tsc --noEmit
   npm run test:coverage  # >= 80%
   npm run build
   ```
3. Add tests for every change in behaviour.
4. **New dependencies** need a reason in the PR description. Prefer the standard library. Changes to
   `.github/`, `Dockerfile`, `go.mod`, `package.json` or `scripts/` are reviewed with extra care and may be
   declined if they aren't essential.

**AI-assisted code is welcome**, but you're responsible for it: you must understand and be able to explain
every line. Large, unreviewed, AI-generated PRs will be closed.

## How PRs are reviewed
A PR can only be merged when **all** of these are true (enforced by GitHub branch rules on `main`):
1. **CI is green**: backend tests (≥ 80% coverage, `gofmt`, `go vet`), frontend tests (≥ 80% coverage, type check,
   build), the data check (`validate_custom.py` + a test import), the Docker image smoke test, and the PR checks
   (no generated or oversized files, no new vulnerable dependencies).
2. **A maintainer has approved it** ([CODEOWNERS](.github/CODEOWNERS)). Pushing new commits after approval
   resets the approval.
3. **All review conversations are resolved.**

For first-time contributors, a maintainer has to approve CI before it runs. That's normal; we'll get to it.

Reviewers check that the change: does what the PR says and nothing else; follows AGENTS.md; has tests; adds no
needless dependencies, network calls, tracking or obfuscated code; and, for data, has correct Jyutping, characters,
register and a compatible source. We may ask for changes or close PRs that don't fit the project.

## Reporting security problems
Please don't open a public issue for a security problem. Use GitHub's
[private vulnerability reporting](https://github.com/kaho-dev/saycanto/security/advisories/new) instead.

## Licensing
By contributing you agree that:
- Code contributions are licensed under **MIT**.
- Dictionary data contributions are licensed under **CC BY-SA 4.0**, to remain compatible with CC-CEDICT
  (CC-Canto's CC BY-SA 3.0 permits this later version).
