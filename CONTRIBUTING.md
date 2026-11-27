# Contributing to SayCanto

Thanks for helping build a better English → Cantonese dictionary!

## Adding or Fixing Dictionary Entries
1. Edit `data/custom.json` (each entry includes `id`, `english`, `jyutping`, `traditional`, `simplified`, `partOfSpeech`, `register`, `classifier`, `notes`, and `examples`).
2. Run the importer locally to verify your JSON parses cleanly:
   ```bash
   cd backend
   go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db
   ```
3. Open a Pull Request!

## Code Contributions
Please read `AGENTS.md` before submitting code PRs. All PRs must pass CI with **>= 80% test coverage** on both backend and frontend.

## Data Licensing
- Code contributions are licensed under **MIT**.
- Dictionary data contributions are licensed under **CC BY-SA 4.0** to remain compatible with CC-CEDICT (CC-Canto's CC BY-SA 3.0 permits this later version).
