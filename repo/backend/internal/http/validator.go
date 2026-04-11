package http

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator wires go-playground/validator into Echo's validation lifecycle.
// Once registered via e.Validator, calling c.Validate(&req) after c.Bind enforces
// all `validate:"..."` struct tags at the HTTP boundary.
type CustomValidator struct {
	v *validator.Validate
}

// NewCustomValidator constructs a CustomValidator with a default validator instance.
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{v: validator.New()}
}

// Validate implements echo.Validator.
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}
