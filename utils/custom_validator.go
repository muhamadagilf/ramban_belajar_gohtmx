package utils

import (
	"regexp"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler/app"
)

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
		return slices.Contains(app.MAJOR, majorStr)
	})

	v.RegisterValidation("oneof_room", func(fl validator.FieldLevel) bool {
		roomStr := fl.Field().String()
		return slices.Contains(app.ROOM, roomStr)
	})

	v.RegisterValidation("nochars", func(fl validator.FieldLevel) bool {
		searchStr := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString(searchStr)
	})

	return &CustomValidator{validator: v}
}
