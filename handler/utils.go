package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

var MAJOR = []string{"TEKNIK INFORMATIKA", "REKAYASA PERANGKAT LUNAK", "AKUNTANSI"}
var ROOM = []string{"TIR1", "TIR2", "RPLR1", "RPLR2", "AKR1", "AKR2"}
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

func studentsQueryParamHandler(c echo.Context, qtx *database.Queries) (Data, error) {
	type StudentsQuery struct {
		Search string `query:"search" validate:"omitempty"`
		Room   string `query:"room" validate:"omitempty,oneof_room"`
		Major  string `query:"major" validate:"omitempty,oneof_major"`
	}

	rawUrlQuery := c.Request().URL.RawQuery
	query := StudentsQuery{}
	if err := c.Bind(&query); err != nil {
		return Data{}, err
	}

	if err := c.Validate(&query); err != nil {
		return Data{}, err
	}

	if rawUrlQuery == "" {
		students, err := qtx.GetStudentAll(c.Request().Context())
		return Data{"Students": students}, err
	}

	if strings.Contains(rawUrlQuery, "search") {
		if _, err := strconv.Atoi(query.Search); err != nil {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%" + query.Search + "%",
					Nim:  "%%",
				})

			return Data{"Students": students}, err
		} else {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%" + query.Search + "%",
					Nim:  "%%",
				})

			return Data{"Students": students}, err
		}
	}

	students, err := qtx.GetStudentsByRoomAndMajor(
		c.Request().Context(),
		database.GetStudentsByRoomAndMajorParams{
			Name:  query.Room,
			Major: query.Major,
		})

	return Data{"Students": students}, err
}
