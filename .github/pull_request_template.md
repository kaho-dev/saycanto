## What does this change?
<!-- One or two sentences. Link the issue it fixes: "Fixes #123". -->

## Type of change
- [ ] Dictionary data (`data/custom.json`)
- [ ] Bug fix
- [ ] New feature
- [ ] Docs only
- [ ] CI / build / dependencies

## Checklist (all PRs)
- [ ] I read [CONTRIBUTING.md](../CONTRIBUTING.md); the PR does one thing and is small enough to review.
- [ ] CI passes (tests, ≥ 80% coverage, formatting, type check, data check).
- [ ] No generated files, databases, secrets or downloaded data are included.
- [ ] I wrote this change myself or can explain every line (AI-assisted is fine; unreviewed AI output is not).

## For dictionary data
- [ ] `python scripts/validate_custom.py` passes.
- [ ] Jyutping, characters and examples come from my own knowledge or a CC BY-SA–compatible source (named below), **not** words.hk or a commercial dictionary.
- [ ] I am a native / fluent Cantonese speaker, **or** I marked the entries as needing native-speaker review.

Source(s):

## For code
- [ ] Follows [AGENTS.md](../AGENTS.md) (feature folders, constructor DI, no SQL in controllers).
- [ ] New or changed behaviour has tests.
- [ ] Explain any new dependency (why it's needed, its licence):
