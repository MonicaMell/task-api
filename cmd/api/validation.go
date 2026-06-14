package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (app *application) failedValidation(w http.ResponseWriter, err error) {
	out := map[string]string{}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			out[strings.ToLower(fe.Field())] = humanizeValidation(fe)
		}
	}
	writeJSON(w, http.StatusBadRequest, map[string]any{"errors": out})
}

func humanizeValidation(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return "invalid value"
	}
}
