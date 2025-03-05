package validate

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ValidationError represents validation failure for a specific field
type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
	Code  string `json:"code"`
}

// ValidationErrors is a custom error type that holds multiple ValidationError
type ValidationErrors []ValidationError

// Error implements the error interface for validationErrors
// It returns a formatted string representation of all validation errors.
// It concatenates the field, message, and code for each validation error
// separated by a comma and space.
// For example: "name: Name is required (required), email: Email is not valid (email)".
// If there are no validation errors, it returns an empty string.
func (ve ValidationErrors) Error() string {
	var errMsgs []string

	for _, err := range ve {
		errMsgs = append(
			errMsgs,
			fmt.Sprintf(
				"%s: %s (%s)",
				err.Field,
				err.Msg,
				err.Code,
			),
		)
	}

	return strings.Join(errMsgs, ", ")
}

// add adds a new validationError to validationErrors
func (ve *ValidationErrors) add(validationError ValidationError) {
	*ve = append(
		*ve,
		validationError,
	)
}

// validate is an initialization of a singleton of the validator package
var validate *validator.Validate
var (
	required               = "required"
	min                    = "min"
	max                    = "max"
	email                  = "email"
	noAllRepeatingCharsTag = "noAllRepeatingChars"
	uuidTag                = "uuid"
	oneof                  = "oneof"
)

func init() {
	validate = validator.New()
	validate.RegisterValidation(noAllRepeatingCharsTag, isNotAllRepeatingChars)
	validate.RegisterValidation(uuidTag, validateUUID)
}

// isNotAllRepeatingChars is a custom validator function that checks if a string
// contain all repeating characters
func isNotAllRepeatingChars(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if len(str) == 0 {
		return false
	}

	uniqueChars := make(map[rune]struct{})
	for _, char := range str {
		uniqueChars[char] = struct{}{}
	}

	return len(uniqueChars) > 1
}

// validateUUID is a custom validator function that checks if a string is a valid
// uuid
func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// StructFields validates the payload against the payload's provided validate
// tags rules.
// It returns a slice of ValidationError if there are validation errors,
// or nil if the payload is valid.
func StructFields[T any](payload T) error {

	if err := validate.Struct(payload); err != nil {
		var validationErrors ValidationErrors

		var validationError ValidationError

		for _, err := range err.(validator.ValidationErrors) {
			validationError.Field = fmt.Sprint(
				strings.ToLower(err.Field()[:1]),
				err.Field()[1:],
			)
			validationError.Code = fmt.Sprintf(
				"%s_%s",
				strings.ToUpper(err.Field()),
				strings.ToUpper(err.Tag()),
			)

			switch err.Tag() {
			case required:
				validationError.Msg = fmt.Sprintf(
					"%s is required",
					validationError.Field,
				)

			case email:
				validationError.Msg = fmt.Sprintf(
					"%s is not a valid email",
					validationError.Field,
				)

			case min:
				validationError.Msg = fmt.Sprintf(
					"%s must be at least %s characters long",
					validationError.Field,
					err.Param(),
				)

			case max:
				validationError.Msg = fmt.Sprintf(
					"%s must be at most %s characters long",
					validationError.Field,
					err.Param(),
				)

			case noAllRepeatingCharsTag:
				validationError.Msg = fmt.Sprintf(
					"%s cannot contain only spaces or repeating characters",
					validationError.Field,
				)

			case oneof:
				validationError.Msg = fmt.Sprintf(
					"%s must be either one of these '%s'",
					validationError.Field,
					err.Param(),
				)

			case uuidTag:
				validationError.Msg = fmt.Sprintf(
					"%s is not a valid id",
					validationError.Field,
				)

			default:
				validationError.Msg = fmt.Sprintf(
					"%s is not valid",
					validationError.Field,
				)
			}

			validationErrors.add(validationError)
		}

		return &validationErrors
	}

	return nil
}
