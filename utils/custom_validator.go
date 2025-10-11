package utils

import (
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler/web"
)

var sqlQueryWord = []string{"select", "delete", "update", "create", "table", "insert"}

// helper to Custom Validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	v.RegisterValidation("oneof_major", func(fl validator.FieldLevel) bool {
		majorStr := fl.Field().String()
		return slices.Contains(web.MAJOR, majorStr)
	})

	v.RegisterValidation("oneof_room", func(fl validator.FieldLevel) bool {
		roomStr := fl.Field().String()
		return slices.Contains(web.ROOM, roomStr)
	})

	v.RegisterValidation("nochars", func(fl validator.FieldLevel) bool {
		searchStr := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString(searchStr)
	})

	v.RegisterValidation("email_constraints", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		return regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(email)
	})

	v.RegisterValidation("phone_constraints", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		return regexp.MustCompile(`^\+?[0-9]{8,15}$`).MatchString(phone)
	})

	v.RegisterValidation("cheeky_sql_inject", func(fl validator.FieldLevel) bool {
		searchVal := fl.Field().String()
		return !slices.Contains(sqlQueryWord, strings.ToLower(searchVal))
	})

	v.RegisterValidation("nip_constraints", func(fl validator.FieldLevel) bool {
		nip := fl.Field().String()
		return regexp.MustCompile(`^[0-9]{16}`).MatchString(nip)
	})

	v.RegisterValidation("dob_constraints", func(fl validator.FieldLevel) bool {
		dob := fl.Field().String()
		return regexp.MustCompile(`^[0-9]{2}\-[a-zA-z]+\-[0-9]{4}$`).MatchString(dob)
	})

	return &CustomValidator{validator: v}
}
