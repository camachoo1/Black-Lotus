package validators_test

import (
	"testing"

	"github.com/go-playground/validator/v10"

	"black-lotus/internal/auth/validators"
)

// This test focuses on testing the validation functions through the validator package
func TestPasswordValidations(t *testing.T) {
	validate := validator.New()
	validators.RegisterCustomValidators(validate)

	type TestPassword struct {
		Password string `validate:"required,containsuppercase,containslowercase,containsnumber,containsspecialchar"`
	}

	testCases := []struct {
		Name              string
		Password          string
		ShouldPass        bool
		FailedValidations []string
	}{
		{
			Name:       "Valid complex password",
			Password:   "Password123!",
			ShouldPass: true,
		},
		{
			Name:              "Missing uppercase",
			Password:          "password123!",
			ShouldPass:        false,
			FailedValidations: []string{"containsuppercase"},
		},
		{
			Name:              "Missing lowercase",
			Password:          "PASSWORD123!",
			ShouldPass:        false,
			FailedValidations: []string{"containslowercase"},
		},
		{
			Name:              "Missing number",
			Password:          "Password!",
			ShouldPass:        false,
			FailedValidations: []string{"containsnumber"},
		},
		{
			Name:              "Missing special char",
			Password:          "Password123",
			ShouldPass:        false,
			FailedValidations: []string{"containsspecialchar"},
		},
		{
			Name:              "Empty password",
			Password:          "",
			ShouldPass:        false,
			FailedValidations: []string{"required"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test := TestPassword{
				Password: tc.Password,
			}

			err := validate.Struct(test)

			if tc.ShouldPass && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}

			if !tc.ShouldPass {
				if err == nil {
					t.Error("Expected validation to fail, but it passed")
				} else {
					// Verify the expected validation errors
					validationErrors, ok := err.(validator.ValidationErrors)
					if !ok {
						t.Errorf("Expected validator.ValidationErrors, got %T", err)
						return
					}

					// Check if all expected validation errors are present
					if len(validationErrors) != len(tc.FailedValidations) {
						t.Errorf("Expected %d validation errors, got %d", len(tc.FailedValidations), len(validationErrors))
					}

					// Create a map of failed tags for easier lookup
					failedTags := make(map[string]bool)
					for _, verr := range validationErrors {
						failedTags[verr.Tag()] = true
					}

					// Check if all expected failed validations are present
					for _, tag := range tc.FailedValidations {
						if !failedTags[tag] {
							t.Errorf("Expected validation '%s' to fail, but it didn't", tag)
						}
					}
				}
			}
		})
	}
}

// Test each validation function individually
func TestValidateContainsUppercase(t *testing.T) {
	validate := validator.New()
	validators.RegisterCustomValidators(validate)

	type TestCase struct {
		Name  string
		Value string
		Valid bool
	}

	testCases := []TestCase{
		{
			Name:  "Contains uppercase",
			Value: "Password123!",
			Valid: true,
		},
		{
			Name:  "No uppercase",
			Value: "password123!",
			Valid: false,
		},
		{
			Name:  "Empty string",
			Value: "",
			Valid: false,
		},
		{
			Name:  "Only uppercase",
			Value: "PASSWORD",
			Valid: true,
		},
		{
			Name:  "Only numbers and special chars",
			Value: "123!@#",
			Valid: false,
		},
	}

	type TestStruct struct {
		Password string `validate:"required,containsuppercase"`
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test := TestStruct{
				Password: tc.Value,
			}

			err := validate.Struct(test)
			if tc.Valid && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}

			if !tc.Valid && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateContainsLowercase(t *testing.T) {
	validate := validator.New()
	validators.RegisterCustomValidators(validate)

	type TestCase struct {
		Name  string
		Value string
		Valid bool
	}

	testCases := []TestCase{
		{
			Name:  "Contains lowercase",
			Value: "Password123!",
			Valid: true,
		},
		{
			Name:  "No lowercase",
			Value: "PASSWORD123!",
			Valid: false,
		},
		{
			Name:  "Empty string",
			Value: "",
			Valid: false,
		},
		{
			Name:  "Only lowercase",
			Value: "password",
			Valid: true,
		},
		{
			Name:  "Only numbers and special chars",
			Value: "123!@#",
			Valid: false,
		},
	}

	type TestStruct struct {
		Password string `validate:"required,containslowercase"`
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test := TestStruct{
				Password: tc.Value,
			}

			err := validate.Struct(test)
			if tc.Valid && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}

			if !tc.Valid && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateContainsNumber(t *testing.T) {
	validate := validator.New()
	validators.RegisterCustomValidators(validate)

	type TestCase struct {
		Name  string
		Value string
		Valid bool
	}

	testCases := []TestCase{
		{
			Name:  "Contains number",
			Value: "Password123!",
			Valid: true,
		},
		{
			Name:  "No number",
			Value: "Password!",
			Valid: false,
		},
		{
			Name:  "Empty string",
			Value: "",
			Valid: false,
		},
		{
			Name:  "Only numbers",
			Value: "12345",
			Valid: true,
		},
		{
			Name:  "Only letters and special chars",
			Value: "Password!",
			Valid: false,
		},
	}

	type TestStruct struct {
		Password string `validate:"required,containsnumber"`
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test := TestStruct{
				Password: tc.Value,
			}

			err := validate.Struct(test)
			if tc.Valid && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}

			if !tc.Valid && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateContainsSpecialChar(t *testing.T) {
	validate := validator.New()
	validators.RegisterCustomValidators(validate)

	type TestCase struct {
		Name  string
		Value string
		Valid bool
	}

	testCases := []TestCase{
		{
			Name:  "Contains special char",
			Value: "Password123!",
			Valid: true,
		},
		{
			Name:  "No special char",
			Value: "Password123",
			Valid: false,
		},
		{
			Name:  "Empty string",
			Value: "",
			Valid: false,
		},
		{
			Name:  "Only special chars",
			Value: "!@#$%^",
			Valid: true,
		},
		{
			Name:  "Only letters and numbers",
			Value: "Password123",
			Valid: false,
		},
		{
			Name:  "Various special chars",
			Value: "Pass!word@123#$%^&*()",
			Valid: true,
		},
	}

	type TestStruct struct {
		Password string `validate:"required,containsspecialchar"`
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test := TestStruct{
				Password: tc.Value,
			}

			err := validate.Struct(test)
			if tc.Valid && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}

			if !tc.Valid && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}
}
