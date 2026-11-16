package handler

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"email must be a valid email address"`
}

type ValidationError struct {
	Status  bool    `json:"status" example:"false"`
	Message string  `json:"message" example:"Bad request"`
	Errors  []Error `json:"errors"`
}

func Validate(err error) []Error {
	slog.Error("bind error", "type", fmt.Sprintf("%T", err), "err", err)
	var ve validator.ValidationErrors
	var errorSlice = make([]Error, 0)

	if errors.As(err, &ve) {
		for _, fe := range ve {
			var field = LowerTrim(fe.Field())
			var tag = fe.Tag()
			var param = LowerTrim(fe.Param())

			switch tag {
			case "gte":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be a greater than " + param})
			case "lte":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be a less than " + param})
			case "oneof":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be one of " + param})
			case "url":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be a valid url"})
			case "email":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be a valid email address."})
			case "required":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " is required."})
			case "required_without":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " is required if " + param + " is not present."})
			case "min":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must be at least " + param + " characters long."})
			case "max":
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " must not exceed " + param + " characters in length."})
			default:
				errorSlice = append(errorSlice, Error{Field: field, Message: field + " failed validation (" + tag + ")"})
			}
		}
		return errorSlice
	}
	return nil
}
