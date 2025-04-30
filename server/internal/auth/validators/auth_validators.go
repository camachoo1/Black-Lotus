// internal/validators/custom_validators.go
package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators adds custom validation functions to the validator
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("containsuppercase", ValidateContainsUppercase)
	v.RegisterValidation("containslowercase", ValidateContainsLowercase)
	v.RegisterValidation("containsnumber", ValidateContainsNumber)
	v.RegisterValidation("containsspecialchar", ValidateContainsSpecialChar)
}

// ValidateContainsUppercase checks if the field contains at least one uppercase letter
func ValidateContainsUppercase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`[A-Z]`).MatchString(value)
}

// ValidateContainsLowercase checks if the field contains at least one lowercase letter
func ValidateContainsLowercase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`[a-z]`).MatchString(value)
}

// ValidateContainsNumber checks if the field contains at least one number
func ValidateContainsNumber(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`[0-9]`).MatchString(value)
}

// ValidateContainsSpecialChar checks if the field contains at least one special character
func ValidateContainsSpecialChar(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(value)
}
