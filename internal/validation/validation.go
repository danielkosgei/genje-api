package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Validator provides validation utilities
type Validator struct {
	errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
	return &Validator{}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) {
	if len(strings.TrimSpace(value)) < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min))
	}
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("must not exceed %d characters", max))
	}
}

// URL validates that a string is a valid URL
func (v *Validator) URL(field, value string) {
	if value == "" {
		return // Skip validation for empty values
	}
	
	if _, err := url.ParseRequestURI(value); err != nil {
		v.AddError(field, "must be a valid URL")
	}
}

// Email validates email format
func (v *Validator) Email(field, value string) {
	if value == "" {
		return // Skip validation for empty values
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.AddError(field, "must be a valid email address")
	}
}

// In validates that a value is in a list of allowed values
func (v *Validator) In(field, value string, allowed []string) {
	if value == "" {
		return // Skip validation for empty values
	}
	
	for _, item := range allowed {
		if value == item {
			return
		}
	}
	
	v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
}

// Range validates that an integer is within a range
func (v *Validator) Range(field string, value, min, max int) {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("must be between %d and %d", min, max))
	}
}

// Positive validates that an integer is positive
func (v *Validator) Positive(field string, value int) {
	if value <= 0 {
		v.AddError(field, "must be positive")
	}
}