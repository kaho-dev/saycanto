# SayCanto Engineering Rules

All contributors and AI agents working in this repository must follow these architectural and quality rules.

## 1. Architecture & Dependency Injection (Backend)
- **Feature-based MVC:** Organize code by feature under `backend/internal/<feature>/` (`model.go`, `repository.go`, `sqlite_repo.go`, `service.go`, `controller.go`).
- **Strict Layering:**
  - **Controller:** HTTP parsing, validation, and JSON serialization via `internal/shared/httputil`. No SQL or business logic.
  - **Service:** Business logic, search normalization, and ranking. Depends *only* on `Repository` interfaces.
  - **Repository:** Database queries and row mapping. Must satisfy the feature's `Repository` interface.
- **Constructor DI:** Pass all dependencies explicitly via `New...(...)` constructors and wire them in `cmd/api/main.go`. Never use global database handles or package-level singletons.

## 2. SOLID & DRY (Backend & Frontend)
- **Single Responsibility:** Keep functions and components small and focused.
- **Dependency Inversion:** Depend on interfaces (Go) or typed client abstractions (TS), not concrete implementations.
- **DRY:**
  - Backend: Reuse `internal/shared/httputil` for JSON responses, error payloads, and CORS.
  - Frontend: Reuse `src/shared/` for API fetching, UI primitives, and formatting helpers. Keep feature code under `src/features/<feature>/`.

## 3. Testing & Coverage (>= 80% Gate)
- Every feature package in `backend/internal/...` must have unit tests using interface mocks (for `Service` and `Controller`) and `:memory:` SQLite (for `sqlite_repo`).
- Frontend features and shared utilities must have Vitest + React Testing Library unit tests.
- **Minimum Coverage:** Both backend (`./internal/...`) and frontend (`src/`) must maintain **>= 80%** test coverage, enforced by CI.
