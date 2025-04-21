package errors

import (
	"fmt"
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestErrorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() error
		expected error
	}{
		{"ErrNotFound", ErrNotFound, errNotFound},
		{"ErrAlreadyExists", ErrAlreadyExists, errAlreadyExists},
		{"ErrInvalidCredentials", ErrInvalidCredentials, errInvalidCredentials},
		{"ErrWrongSigningMethod", ErrWrongSigningMethod, errWrongSigningMethod},
		{"ErrCityNotAllowed", ErrCityNotAllowed, errCityNotAllowed},
		{"ErrActiveReceptionExists", ErrActiveReceptionExists, errActiveReceptionExists},
		{"ErrNoActiveReception", ErrNoActiveReception, errNoActiveReception},
		{"ErrProductTypeNotAllowed", ErrProductTypeNotAllowed, errProductTypeNotAllowed},
		{"ErrNoProduct", ErrNoProduct, errNoProduct},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			assert.Equal(t, tt.expected, err)
			assert.Equal(t, tt.expected.Error(), err.Error())
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		field    string
		expected string
	}{
		{"Required field", "required", "Name", "field Name is a required field"},
		{"UUID field", "uuid", "ID", "field ID is not a valid uuid"},
		{"Email field", "email", "Email", "field Email is not a valid email"},
		{"Datetime field", "datetime", "CreatedAt", "field CreatedAt is not a valid datetime"},
		{"Oneof field", "oneof", "Status", "field Status is not a valid: value is unknown"},
		{"Default case", "unknown", "Field", "field Field is not valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidationErrors{
				&fieldError{field: tt.field, tag: tt.tag},
			}
			result := ValidationError(errs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationError_MultipleErrors(t *testing.T) {
	errs := validator.ValidationErrors{
		&fieldError{field: "Name", tag: "required"},
		&fieldError{field: "Email", tag: "email"},
		&fieldError{field: "ID", tag: "uuid"},
	}

	expected := "field Name is a required field, field Email is not a valid email, field ID is not a valid uuid"
	result := ValidationError(errs)
	assert.Equal(t, expected, result)
}

type fieldError struct {
	field string
	tag   string
}

func (f *fieldError) Tag() string       { return f.tag }
func (f *fieldError) ActualTag() string { return f.tag }
func (f *fieldError) Field() string     { return f.field }
func (f *fieldError) Error() string     { return fmt.Sprintf("field %s is not valid", f.field) }

func (f *fieldError) Namespace() string                 { return "" }
func (f *fieldError) StructNamespace() string           { return "" }
func (f *fieldError) StructField() string               { return "" }
func (f *fieldError) Value() interface{}                { return nil }
func (f *fieldError) Param() string                     { return "" }
func (f *fieldError) Kind() reflect.Kind                { return reflect.String }
func (f *fieldError) Type() reflect.Type                { return reflect.TypeOf("") }
func (f *fieldError) Translate(ut ut.Translator) string { return "" }
