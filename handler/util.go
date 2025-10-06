package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type dbFunc = func(q *database.Queries) error

func WithTX(ctx context.Context, db *sql.DB, q *database.Queries, fn dbFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()
	qtx := q.WithTx(tx)

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}

func IsLastModifiedValid(modifiedSince string, lastModified time.Time) bool {
	if modifiedSince == "" {
		return false
	}

	t, err := time.Parse(http.TimeFormat, modifiedSince)

	return err == nil && !lastModified.After(t)
}
