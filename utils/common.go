package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	MAJOR = []string{"TEKNIK INFORMATIKA", "REKAYASA PERANGKAT LUNAK", "AKUNTANSI"}
	ROOM  = []string{"TIR1", "TIR2", "RPLR1", "RPLR2", "AKR1", "AKR2"}
)

const (
	USER_ROLE_STUDENT   = "student"
	USER_ROLE_ADMIN     = "admin"
	USER_ROLE_TEACHER   = "teacher"
	USER_ROLE_SUPERUSER = "superuser"

	// error message
	ERROR_USER_UNAUTHENTICATED     = "You're Not Authenticated, Cannot Access !!!"
	ERROR_USER_UNAUTHORIZED        = "Permission Denied: user unauthorized, not allowed to access"
	ERROR_FAILED_AUTHENTICATION    = "Authentication Failed, input the valid email & password"
	ERROR_INVALID_NIP              = "error: invalid nomer induk pengguna (nip), please check your birthdate/nip"
	ERROR_INVALID_CONFIRM_PASSWORD = "error: your confirmation password is invalid"
	ERROR_INVALID_INPUT_DATA       = "error: invalid input data. please check again and follow the proper data format"
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

func ValidationErrorMsg(errMsg string) string {
	errMsg = strings.ToLower(errMsg)

	if strings.Contains(errMsg, `name_constraints`) {
		return "error: invalid name, violates name_constraints"
	}

	if strings.Contains(errMsg, `nip_constraints`) {
		return "error: invalid nomer induk pengguna, violates nip_constraints"
	}

	if strings.Contains(errMsg, `phone_constraints`) {
		return "error: invalid phone number, please input the valid number"
	}

	if strings.Contains(errMsg, `email_constraints`) {
		return "error: invalid email address, please input the valid address"
	}

	if strings.Contains(errMsg, `dob_constraints`) {
		return "error: wrong format date of birth, please input the right format"
	}

	if strings.Contains(errMsg, `password_constraints`) {
		return ERROR_FAILED_AUTHENTICATION
	}

	return errMsg
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

func InternalServerErrorMessage(debugCode, errMsg string) string {
	return fmt.Sprintf("500 Internal Server Error; Please Contact Support with CODE:%s. \n%v", debugCode, errMsg)
}
