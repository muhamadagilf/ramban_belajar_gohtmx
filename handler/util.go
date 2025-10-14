// Package handler
package handler

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"golang.org/x/crypto/bcrypt"
)

const (
	USER_ROLE_STUDENT = "student"
	USER_ROLE_ADMIN   = "admin"
	USER_ROLE_TEACHER = "teacher"

	// error message
	ERROR_USER_UNAUTHENTICATED     = "You're UnAuthenticated User, Cannot Access !!!"
	ERROR_INVALID_NIP              = "error: invalid nomer induk pengguna (nip), please check your birthdate/nip"
	ERROR_INVALID_CONFIRM_PASSWORD = "error: your confirmation password is invalid"
)

type dbFunc = func(q *database.Queries) error

var DOBLayout = "02-January-2006"

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

func parseBDay(b string) string {
	// months := []string{
	// 	"January",
	// 	"February",
	// 	"March",
	// 	"April",
	// 	"May",
	// 	"June",
	// 	"July",
	// 	"August",
	// 	"September",
	// 	"Oktober",
	// 	"November",
	// 	"December",
	// }

	year := strings.Split(b, "-")[0][2:]
	month := strings.Split(b, "-")[1]
	day := strings.Split(b, "-")[2]

	return day + month + year
}

func IsNIPValid(nip, birthday string) bool {
	parsedDate := parseBDay(birthday)
	log.Println(parsedDate)
	return strings.Contains(nip, parsedDate)
}

func SubmissionErrorMsg(err string) string {
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Use cost 14 for strong security
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
