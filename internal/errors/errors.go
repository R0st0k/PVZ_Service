package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	errNotFound      = errors.New("not found")
	errAlreadyExists = errors.New("already exists")

	errCityNotAllowed        = errors.New("city not allowed")
	errProductTypeNotAllowed = errors.New("product type not allowed")
	errActiveReceptionExists = errors.New("active reception already exists")
	errNoActiveReception     = errors.New("no active reception")
	errNoProduct             = errors.New("no product")

	errInvalidCredentials = errors.New("invalid credentials")
	errWrongSigningMethod = errors.New("unexpected signing method")
)

func ErrNotFound() error              { return errNotFound }
func ErrAlreadyExists() error         { return errAlreadyExists }
func ErrInvalidCredentials() error    { return errInvalidCredentials }
func ErrWrongSigningMethod() error    { return errWrongSigningMethod }
func ErrCityNotAllowed() error        { return errCityNotAllowed }
func ErrActiveReceptionExists() error { return errActiveReceptionExists }
func ErrNoActiveReception() error     { return errNoActiveReception }
func ErrProductTypeNotAllowed() error { return errProductTypeNotAllowed }
func ErrNoProduct() error             { return errNoProduct }

func ValidationError(errs validator.ValidationErrors) string {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "uuid":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid uuid", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid email", err.Field()))
		case "datetime":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid datetime", err.Field()))
		case "oneof":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid: value is unknown", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return strings.Join(errMsgs, ", ")
}
