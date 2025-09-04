package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func init() {
	v = validator.New()
	// ISO4217: три буквы (упрощённая проверка)
	_ = v.RegisterValidation("iso4217", func(fl validator.FieldLevel) bool {
		s := strings.TrimSpace(fl.Field().String())
		if len(s) != 3 {
			return false
		}
		for _, r := range s {
			if r < 'A' || r > 'Z' {
				return false
			}
		}
		return true
	})
}

func ValidateStruct(s any) error {
	return v.Struct(s)
}
