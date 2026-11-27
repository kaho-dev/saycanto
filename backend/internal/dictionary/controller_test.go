package dictionary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockService struct {
	searchFn    func(ctx context.Context, q SearchQuery) (SearchResult, error)
	getByIDFn   func(ctx context.Context, id string) (*Entry, error)
	listWordsFn func(ctx context.Context, q ListQuery) (ListResult, error)
}

func (m *mockService) Search(ctx context.Context, q SearchQuery) (SearchResult, error) {
	return m.searchFn(ctx, q)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Entry, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockService) ListWords(ctx context.Context, q ListQuery) (ListResult, error) {
	return m.listWordsFn(ctx, q)
}

func TestControllerSearch(t *testing.T) {
	t.Run("returns 200 with search results", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			searchFn: func(ctx context.Context, q SearchQuery) (SearchResult, error) {
				return SearchResult{
					Query:   q.Query,
					Total:   1,
					Entries: []Entry{{ID: "m4-goi1", Jyutping: "m4 goi1"}},
				}, nil
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/search?q=thank&limit=5&register=spoken", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var res SearchResult
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if res.Total != 1 || res.Entries[0].ID != "m4-goi1" {
			t.Fatalf("unexpected response: %+v", res)
		}
	})

	t.Run("returns 400 on ErrInvalidInput", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			searchFn: func(ctx context.Context, q SearchQuery) (SearchResult, error) {
				return SearchResult{}, ErrInvalidInput
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/search?q=test&register=bad", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns 500 on unexpected error", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			searchFn: func(ctx context.Context, q SearchQuery) (SearchResult, error) {
				return SearchResult{}, errors.New("unexpected")
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestControllerGetByID(t *testing.T) {
	t.Run("returns 200 when found", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			getByIDFn: func(ctx context.Context, id string) (*Entry, error) {
				return &Entry{ID: id, Jyutping: "seoi2"}, nil
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words/seoi2", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			getByIDFn: func(ctx context.Context, id string) (*Entry, error) {
				return nil, ErrNotFound
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words/missing", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("returns 400 on ErrInvalidInput", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			getByIDFn: func(ctx context.Context, id string) (*Entry, error) {
				return nil, ErrInvalidInput
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words/bad", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}

func TestControllerListWords(t *testing.T) {
	t.Run("returns 200 with paginated words", func(t *testing.T) {
		r := chi.NewRouter()
		var captured ListQuery
		NewController(&mockService{
			listWordsFn: func(ctx context.Context, q ListQuery) (ListResult, error) {
				captured = q
				return ListResult{
					Total:   2,
					Limit:   q.Limit,
					Offset:  q.Offset,
					Entries: []Entry{{ID: "m4-goi1", Jyutping: "m4 goi1"}},
				}, nil
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words?limit=10&offset=5&register=spoken", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if captured.Limit != 10 || captured.Offset != 5 || captured.Register != "spoken" {
			t.Fatalf("unexpected query params: %+v", captured)
		}
		var res ListResult
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if res.Total != 2 || len(res.Entries) != 1 || res.Entries[0].ID != "m4-goi1" {
			t.Fatalf("unexpected response: %+v", res)
		}
	})

	t.Run("returns 400 on ErrInvalidInput", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			listWordsFn: func(ctx context.Context, q ListQuery) (ListResult, error) {
				return ListResult{}, ErrInvalidInput
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words?register=bad", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns 500 on unexpected error", func(t *testing.T) {
		r := chi.NewRouter()
		NewController(&mockService{
			listWordsFn: func(ctx context.Context, q ListQuery) (ListResult, error) {
				return ListResult{}, errors.New("unexpected")
			},
		}).RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/words", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}
