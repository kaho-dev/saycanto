package importer

import (
	"context"
	"database/sql"

	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
)

// BuildDatabase initializes the SQLite FTS5 schema and populates all entries.
func BuildDatabase(ctx context.Context, db *sql.DB, entries []dictionary.Entry) error {
	if err := dictionary.InitSchema(db); err != nil {
		return err
	}
	return dictionary.InsertEntries(ctx, db, entries)
}
