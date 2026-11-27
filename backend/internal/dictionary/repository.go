package dictionary

import "context"

// Repository defines persistence operations for dictionary entries (DIP).
type Repository interface {
	Search(ctx context.Context, q SearchQuery) ([]Entry, error)
	// CountTypes counts all matches of q.Query (and q.Register) per word type; q.WordType is ignored.
	CountTypes(ctx context.Context, q SearchQuery) (TypeCounts, error)
	GetByID(ctx context.Context, id string) (*Entry, error)
	List(ctx context.Context, q ListQuery) ([]Entry, int, error)
}
