package migration

import (
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const MigrationName = "00001_init.up.sql"

//go:embed 00001_init.up.sql
var schemaSQL string

func Apply(db *sqlx.DB) (applied bool, err error) {
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = 'events'
		);
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	if exists {
		return false, nil
	}

	_, err = db.Exec(schemaSQL)
	if err != nil {
		return false, fmt.Errorf("failed to apply migration: %w", err)
	}

	return true, nil
}
