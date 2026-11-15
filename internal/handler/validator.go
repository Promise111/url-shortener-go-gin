package handler

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

func Validate(err error) {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		for _, fe := range ve {
			slog.Error("Validation Error", "field", fe.Field(), "tag", fe.Tag(),"param",fe.Param())
		}
	}else {
		slog.Error("Malformed JSON")
	}
}