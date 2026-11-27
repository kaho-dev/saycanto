package dictionary

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	searchFn     func(ctx context.Context, q SearchQuery) ([]Entry, error)
	getByIDFn    func(ctx context.Context, id string) (*Entry, error)
	listFn       func(ctx context.Context, q ListQuery) ([]Entry, int, error)
	countFn      func(ctx context.Context, q SearchQuery) (TypeCounts, error)
	defaultTotal int
}

// CountTypes defaults to TotalFromSearch, which the service treats like any count; tests that
// check totals set countFn.
func (m *mockRepo) CountTypes(ctx context.Context, q SearchQuery) (TypeCounts, error) {
	if m.countFn != nil {
		return m.countFn(ctx, q)
	}
	return TypeCounts{Total: m.defaultTotal}, nil
}

func (m *mockRepo) Search(ctx context.Context, q SearchQuery) ([]Entry, error) {
	return m.searchFn(ctx, q)
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Entry, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) List(ctx context.Context, q ListQuery) ([]Entry, int, error) {
	return m.listFn(ctx, q)
}

func TestServiceSearch(t *testing.T) {
	t.Run("empty query returns empty result without calling repo", func(t *testing.T) {
		svc := NewService(&mockRepo{
			searchFn: func(ctx context.Context, q SearchQuery) ([]Entry, error) {
				t.Fatal("repo should not be called on empty query")
				return nil, nil
			},
		})
		res, err := svc.Search(context.Background(), SearchQuery{Query: "   "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 0 || len(res.Entries) != 0 {
			t.Fatalf("expected empty result, got %+v", res)
		}
	})

	t.Run("normalizes query, clamps limit, and validates register", func(t *testing.T) {
		var captured SearchQuery
		svc := NewService(&mockRepo{
			searchFn: func(ctx context.Context, q SearchQuery) ([]Entry, error) {
				captured = q
				return []Entry{{ID: "m4-goi1", Jyutping: "m4 goi1"}}, nil
			},
			defaultTotal: 1,
		})
		res, err := svc.Search(context.Background(), SearchQuery{
			Query:    "  thank you ",
			Register: " SPOKEN ",
			Limit:    999,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Query != "thank you" || captured.Register != "spoken" || captured.Limit != maxSearchLimit {
			t.Fatalf("unexpected normalized query: %+v", captured)
		}
		if res.Total != 1 || res.Entries[0].ID != "m4-goi1" {
			t.Fatalf("unexpected result: %+v", res)
		}
	})

	t.Run("defaults limit when <= 0 and handles nil slice from repo", func(t *testing.T) {
		var captured SearchQuery
		svc := NewService(&mockRepo{
			searchFn: func(ctx context.Context, q SearchQuery) ([]Entry, error) {
				captured = q
				return nil, nil
			},
		})
		res, err := svc.Search(context.Background(), SearchQuery{Query: "water", Limit: -5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Limit != defaultSearchLimit {
			t.Fatalf("expected default limit %d, got %d", defaultSearchLimit, captured.Limit)
		}
		if res.Entries == nil {
			t.Fatalf("expected non-nil empty slice")
		}
	})

	t.Run("rejects invalid register", func(t *testing.T) {
		svc := NewService(&mockRepo{})
		_, err := svc.Search(context.Background(), SearchQuery{Query: "water", Register: "invalid"})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repoErr := errors.New("db down")
		svc := NewService(&mockRepo{
			searchFn: func(ctx context.Context, q SearchQuery) ([]Entry, error) {
				return nil, repoErr
			},
		})
		_, err := svc.Search(context.Background(), SearchQuery{Query: "water"})
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestServiceGetByID(t *testing.T) {
	t.Run("rejects blank id", func(t *testing.T) {
		svc := NewService(&mockRepo{})
		_, err := svc.GetByID(context.Background(), "  ")
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("trims id and delegates to repo", func(t *testing.T) {
		svc := NewService(&mockRepo{
			getByIDFn: func(ctx context.Context, id string) (*Entry, error) {
				if id != "seoi2" {
					t.Fatalf("expected trimmed id 'seoi2', got %q", id)
				}
				return &Entry{ID: id, Jyutping: "seoi2"}, nil
			},
		})
		entry, err := svc.GetByID(context.Background(), " seoi2 ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.ID != "seoi2" {
			t.Fatalf("unexpected entry: %+v", entry)
		}
	})
}

func TestServiceListWords(t *testing.T) {
	t.Run("clamps limit and negative offset, normalizes register", func(t *testing.T) {
		var captured ListQuery
		svc := NewService(&mockRepo{
			listFn: func(ctx context.Context, q ListQuery) ([]Entry, int, error) {
				captured = q
				return []Entry{{ID: "m4-goi1", Jyutping: "m4 goi1"}}, 10, nil
			},
		})
		res, err := svc.ListWords(context.Background(), ListQuery{
			Register: " SPOKEN ",
			Limit:    999,
			Offset:   -5,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Limit != maxSearchLimit || captured.Offset != 0 || captured.Register != "spoken" {
			t.Fatalf("unexpected normalized query: %+v", captured)
		}
		if res.Total != 10 || res.Limit != maxSearchLimit || res.Offset != 0 || len(res.Entries) != 1 {
			t.Fatalf("unexpected result: %+v", res)
		}
	})

	t.Run("defaults limit when <= 0 and converts nil slice to empty slice", func(t *testing.T) {
		var captured ListQuery
		svc := NewService(&mockRepo{
			listFn: func(ctx context.Context, q ListQuery) ([]Entry, int, error) {
				captured = q
				return nil, 0, nil
			},
		})
		res, err := svc.ListWords(context.Background(), ListQuery{Limit: 0, Offset: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Limit != defaultSearchLimit || captured.Offset != 2 {
			t.Fatalf("unexpected query: %+v", captured)
		}
		if res.Entries == nil || len(res.Entries) != 0 {
			t.Fatalf("expected non-nil empty slice, got %+v", res.Entries)
		}
	})

	t.Run("rejects invalid register", func(t *testing.T) {
		svc := NewService(&mockRepo{})
		_, err := svc.ListWords(context.Background(), ListQuery{Register: "invalid"})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repoErr := errors.New("db down")
		svc := NewService(&mockRepo{
			listFn: func(ctx context.Context, q ListQuery) ([]Entry, int, error) {
				return nil, 0, repoErr
			},
		})
		_, err := svc.ListWords(context.Background(), ListQuery{Limit: 10})
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestServiceSearchWordTypes(t *testing.T) {
	var gotSearch, gotCount SearchQuery
	repo := &mockRepo{
		searchFn: func(ctx context.Context, q SearchQuery) ([]Entry, error) {
			gotSearch = q
			return []Entry{{ID: "daan6-hai6"}}, nil
		},
		countFn: func(ctx context.Context, q SearchQuery) (TypeCounts, error) {
			gotCount = q
			return TypeCounts{Total: 40, ByType: map[string]int{"conjunction": 3, "noun": 5}, Untyped: 30}, nil
		},
	}
	svc := NewService(repo)

	res, err := svc.Search(context.Background(), SearchQuery{Query: "but", WordType: " Conjunction "})
	if err != nil {
		t.Fatal(err)
	}
	if gotSearch.WordType != "conjunction" || gotCount.WordType != "" {
		t.Errorf("search type %q, count type %q", gotSearch.WordType, gotCount.WordType)
	}
	if res.Total != 3 || res.TypeCounts["noun"] != 5 || res.Untyped != 30 {
		t.Errorf("result = %+v", res)
	}

	res, _ = svc.Search(context.Background(), SearchQuery{Query: "but"})
	if res.Total != 40 {
		t.Errorf("unfiltered total = %d, want 40", res.Total)
	}

	if _, err := svc.Search(context.Background(), SearchQuery{Query: "but", WordType: "gerund"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput, got %v", err)
	}
	if _, err := svc.ListWords(context.Background(), ListQuery{WordType: "gerund"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("list: want ErrInvalidInput, got %v", err)
	}

	repo.countFn = func(ctx context.Context, q SearchQuery) (TypeCounts, error) { return TypeCounts{}, errors.New("boom") }
	if _, err := svc.Search(context.Background(), SearchQuery{Query: "but"}); err == nil {
		t.Error("want count error")
	}
}
