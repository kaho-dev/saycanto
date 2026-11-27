# Maintaining SayCanto: protecting `main`

Pushing to `main` deploys production (Cloud Run + Vercel), so `main` must only change through reviewed PRs.
The files in the repo ([CODEOWNERS](../.github/CODEOWNERS), the PR template, CI) only do their job once these
GitHub settings are on. Outside contributors can never merge on a public repo; these settings also stop
**collaborators** from merging, and stop anyone (you included) from pushing untested code straight to `main`.

## 1. Branch ruleset for `main`
GitHub → repo **Settings → Rules → Rulesets → New branch ruleset**:
- Name `protect-main`, Enforcement **Active**, Target branches **Include default branch**.
- Bypass list: **Repository admin** (you), set to **"For pull requests only"**, so you can still merge your own
  PRs while you're the only reviewer (GitHub doesn't let you approve your own PR) but can't push directly.
- Turn on:
  - **Restrict deletions** and **Block force pushes**.
  - **Require a pull request before merging**: 1 required approval; **Dismiss stale approvals when new commits
    are pushed**; **Require review from Code Owners**; **Require conversation resolution before merging**.
  - **Require status checks to pass**, with **Require branches to be up to date before merging**, and add:
    `Backend`, `Frontend`, `Data`, `Docker image`, `PR checks`.
    (A check only shows up in the list after it has run once, so open a test PR first.)

When a second trusted maintainer joins, remove yourself from the bypass list so every PR, yours included, needs
someone else's approval.

## 2. Actions settings
**Settings → Actions → General**:
- *Fork pull request workflows from outside collaborators*: **Require approval for all outside collaborators**.
  Read a fork PR's diff, especially `.github/`, `Dockerfile`, `scripts/`, before clicking "Approve and run".
- *Workflow permissions*: **Read repository contents and packages permissions**; leave "Allow GitHub Actions to
  create and approve pull requests" **off**.

## 3. Security settings
**Settings → Code security**: turn on **Dependency graph** (needed by the `PR checks` job), **Dependabot alerts**,
**Secret scanning** with **Push protection**, and **Private vulnerability reporting**.

## 4. Deploy integrations
- **Vercel:** keep **Git Fork Protection** on (Project → Settings → Git), so a fork PR can't build with your
  environment variables until you allow it.
- **Cloud Build:** the trigger must stay on **push to `^main$`** only, never "pull request".

## Reviewing a PR
1. Read the whole diff, not just the description. Be most suspicious of changes to `.github/`, `Dockerfile`,
   `go.mod`/`go.sum`, `package.json`/`package-lock.json`, `scripts/` and anything that makes network calls.
2. New dependency? Check it's needed, maintained and MIT/BSD/Apache-compatible.
3. Data PRs: check Jyutping and characters against a trusted source, the register, and that the stated source
   is CC BY-SA–compatible. Unsure? Add the entries to the native-speaker review list in `docs/STATUS.md`.
4. Prefer **Squash and merge** so `main` gets one clean commit per PR.
