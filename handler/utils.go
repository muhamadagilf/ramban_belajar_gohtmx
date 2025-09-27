package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

var MAJOR = []string{"TEKNIK INFORMATIKA", "REKAYASA PERANGKAT LUNAK", "AKUNTANSI"}
var YEAR = time.Now().Year()

type dbFunc = func(q *database.Queries) error

type Server struct {
	Queries *database.Queries
	DB      *sql.DB
}

func GetServerConfig() (Server, error) {

	dbURL := os.Getenv("db_url")
	if dbURL == "" {
		return Server{}, fmt.Errorf("error: cannot find db_url in the environment")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return Server{}, err
	}

	return Server{Queries: database.New(conn), DB: conn}, nil

}

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

func submissionErrorMsg(err string) string {

	err = strings.ToLower(err)

	if strings.Contains(err, `violates check constraint "students_phone_number_check"`) {
		return "error: invalid phone number, please input the correct number"
	}

	if strings.Contains(err, `violates unique constraint "students_phone_number_key"`) {
		return "error: cannot input phone number, number already exists"
	}

	if strings.Contains(err, `violates unique constraint "students_email_key"`) {
		return "error: cannot input email, email already exists"
	}

	if strings.Contains(err, `violates unique constraint "students_nip_key"`) {
		return "error: cannot input NIP, NIP already exists"
	}

	return err
}

func isLastModifiedValid(modifiedSince string, lastModified time.Time) bool {
	if modifiedSince == "" {
		return false
	}

	t, err := time.Parse(http.TimeFormat, modifiedSince)

	return err == nil && !lastModified.After(t)
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}
