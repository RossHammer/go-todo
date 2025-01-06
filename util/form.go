package util

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type FormReader struct {
	decoder    *form.Decoder
	validate   *validator.Validate
	translator ut.Translator
}

type FieldError struct {
	Field   string
	Message string
}

func NewFormReader() *FormReader {
	validate := validator.New(validator.WithRequiredStructEnabled())
	en := en.New()
	uni := ut.New(en)
	trans := uni.GetFallback()
	err := en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(err)
	}

	return &FormReader{
		decoder:    form.NewDecoder(),
		validate:   validate,
		translator: trans,
	}
}

func (f *FormReader) ReadForm(v interface{}, r *http.Request) ([]FieldError, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("error parsing form: %w", err)
	}

	if err := f.decoder.Decode(v, r.PostForm); err != nil {
		return nil, fmt.Errorf("error decoding form: %w", err)
	}

	err := f.validate.StructCtx(r.Context(), v)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errors := make([]FieldError, len(validationErrors))
		for i, err := range validationErrors {
			errors[i] = FieldError{
				Field:   err.Field(),
				Message: err.Translate(f.translator),
			}
		}
		return errors, nil
	} else if err != nil {
		return nil, fmt.Errorf("error validating form: %w", err)
	}
	return nil, nil
}
