package dbinit

import (
	"database/sql"
	"fmt"
	"sort"

	"carnet/migrations"
)

// Apply runs every embedded .sql migration file, in filename order.
// Files are written as idempotent statements (CREATE TABLE IF NOT EXISTS, etc.)
// so this is safe to run on every startup.
func Apply(db *sql.DB) error {
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("lecture des migrations: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		content, err := migrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("lecture de %s: %w", name, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("execution de %s: %w", name, err)
		}
	}
	return nil
}
