// internal/common/validation/password.go
package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Register custom validators for password requirements
func RegisterPasswordValidators(v *validator.Validate) {
	// Contains uppercase letter
	_ = v.RegisterValidation("containsuppercase", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`[A-Z]`).MatchString(fl.Field().String())
	})

	// Contains lowercase letter
	_ = v.RegisterValidation("containslowercase", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`[a-z]`).MatchString(fl.Field().String())
	})

	// Contains number
	_ = v.RegisterValidation("containsnumber", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`[0-9]`).MatchString(fl.Field().String())
	})

	// Contains special character
	_ = v.RegisterValidation("containsspecialchar", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(fl.Field().String())
	})
}
