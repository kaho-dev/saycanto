# SayCanto

An open-source **English → Cantonese** dictionary built with **Jyutping** front-and-center, spoken vs. formal register tags, classifiers, and natural example sentences.

## Tech Stack
- **Backend:** Go (`chi` + embedded SQLite FTS5 via `modernc.org/sqlite`), MVC + feature-based clean architecture with manual constructor DI.
- **Frontend:** Next.js (App Router, TypeScript, Tailwind CSS) + Vitest & React Testing Library.
- **Data:** Curated English-first JSON overlay (`data/custom.json`) + importers for CC-Canto and CC-CEDICT with Cantonese readings (`CC BY-SA 4.0`).

## Quick Start

### 1. Build the Dictionary Database & Start Backend
```bash
cd backend
go run ./cmd/importer -custom ../data/custom.json -out ./dictionary.db
go run ./cmd/api -db ./dictionary.db -port 8080
```
*(For the full dictionary, add `-cccanto`, `-cedict` and `-cedict-readings`; see `data/README.md`.)*

### 2. Start Frontend
```bash
cd frontend
npm install
npm run dev
```
Open http://localhost:3000.

## Testing & Coverage (>= 80% Required)
```bash
# Backend
cd backend && go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out

# Frontend
cd frontend && npm run test:coverage
```

## Licensing
- **Source Code:** MIT License (see `LICENSE`)
- **Dictionary Data:** **CC BY-SA 4.0**. Built from CC-Canto and CC-CEDICT Cantonese readings (© Pleco Software, CC BY-SA 3.0) and CC-CEDICT (MDBG and contributors, CC BY-SA 4.0); see `data/README.md`.
