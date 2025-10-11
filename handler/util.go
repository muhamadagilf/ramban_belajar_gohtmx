package handler

import (
	"context"
	"database/sql"
	"net/http"
	"slices"
	"strconv"
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
	ERROR_USER_UNAUTHENTICATED = "You're UnAuthenticated User, Cannot Access !!!"
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
	months := []string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"Oktober",
		"November",
		"December",
	}

	birthSplit := strings.Split(b, "-")
	day := birthSplit[0]
	month := birthSplit[1]
	year := birthSplit[2][2:]

	mn := slices.Index(months, month) + 1
	month = "0" + strconv.Itoa(mn)

	return day + month + year
}

func IsNIPValid(nip, birthday string) bool {
	parsedDate := parseBDay(birthday)
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

// A cost of 12-14 is commonly recommended for production.
func HashPassword(password string) (string, error) {
	// Convert password to byte slice
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Use cost 14 for strong security
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash verifies if the given password matches the stored hash.
func CheckPasswordHash(password, hash string) bool {
	// Compare the hash with the password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
