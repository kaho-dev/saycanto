package dictionary

import (
	"context"
	"strings"
)

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 50
)

// Service defines business operations for the dictionary feature.
type Service interface {
	Search(ctx context.Context, q SearchQuery) (SearchResult, error)
	GetByID(ctx context.Context, id string) (*Entry, error)
	ListWords(ctx context.Context, q ListQuery) (ListResult, error)
}

type serviceImpl struct {
	repo Repository
}

// NewService constructs a dictionary Service with an injected Repository.
func NewService(repo Repository) Service {
	return &serviceImpl{repo: repo}
}

// Search normalizes the query, enforces limits, and delegates to the repository.
func (s *serviceImpl) Search(ctx context.Context, q SearchQuery) (SearchResult, error) {
	trimmed := strings.TrimSpace(q.Query)
	if trimmed == "" {
		return SearchResult{Query: "", Total: 0, Entries: []Entry{}, TypeCounts: map[string]int{}}, nil
	}

	register, err := normalizeRegister(q.Register)
	if err != nil {
		return SearchResult{}, err
	}

	wordType, err := normalizeWordType(q.WordType)
	if err != nil {
		return SearchResult{}, err
	}

	// The results and the per-type counts are independent reads, so run them at the same time.
	type countResult struct {
		counts TypeCounts
		err    error
	}
	countCh := make(chan countResult, 1)
	go func() {
		c, err := s.repo.CountTypes(ctx, SearchQuery{Query: trimmed, Register: register})
		countCh <- countResult{c, err}
	}()

	entries, err := s.repo.Search(ctx, SearchQuery{
		Query:    trimmed,
		Register: register,
		WordType: wordType,
		Limit:    clampLimit(q.Limit),
	})
	cr := <-countCh
	if err != nil {
		return SearchResult{}, err
	}
	if cr.err != nil {
		return SearchResult{}, cr.err
	}
	counts := cr.counts
	if entries == nil {
		entries = []Entry{}
	}
	if counts.ByType == nil {
		counts.ByType = map[string]int{}
	}
	total := counts.Total
	if wordType != "" {
		total = counts.ByType[wordType]
	}

	return SearchResult{
		Query:      trimmed,
		Total:      total,
		Entries:    entries,
		TypeCounts: counts.ByType,
		Untyped:    counts.Untyped,
	}, nil
}

// GetByID validates the ID and fetches a single dictionary entry.
func (s *serviceImpl) GetByID(ctx context.Context, id string) (*Entry, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetByID(ctx, trimmed)
}

// ListWords validates pagination and filter parameters and returns a paginated list of entries.
func (s *serviceImpl) ListWords(ctx context.Context, q ListQuery) (ListResult, error) {
	register, err := normalizeRegister(q.Register)
	if err != nil {
		return ListResult{}, err
	}
	wordType, err := normalizeWordType(q.WordType)
	if err != nil {
		return ListResult{}, err
	}

	limit := clampLimit(q.Limit)
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	entries, total, err := s.repo.List(ctx, ListQuery{
		Register: register,
		WordType: wordType,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return ListResult{}, err
	}
	if entries == nil {
		entries = []Entry{}
	}

	return ListResult{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Entries: entries,
	}, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultSearchLimit
	}
	if limit > maxSearchLimit {
		return maxSearchLimit
	}
	return limit
}

func normalizeWordType(raw string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(raw))
	if t != "" && !IsWordType(t) {
		return "", ErrInvalidInput
	}
	return t, nil
}

func normalizeRegister(raw string) (string, error) {
	register := strings.ToLower(strings.TrimSpace(raw))
	if register != "" && register != "spoken" && register != "formal" && register != "both" {
		return "", ErrInvalidInput
	}
	return register, nil
}
