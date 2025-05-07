package middlewares

import (
	"errors"
	"log"
	"net/http"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
)

// ErrorHandler is a middleware that takes handler that returns an error and
// return a HandlerFunc to create a centralized error handling, logging and etc.
func (mw *middleware) ErrorHandler(h handlerutils.APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			log.Println(err)
			var serverError *servererrors.ServerError

			if errors.As(err, &serverError) {
				switch serverError.StatusCode {
				case http.StatusBadRequest:
					handlerutils.WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusConflict:
					handlerutils.WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusUnprocessableEntity:
					handlerutils.WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)

				case http.StatusUnauthorized:
					handlerutils.WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusForbidden:
					handlerutils.WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				}
			} else {
				handlerutils.WriteErrorJSON(
					w,
					http.StatusInternalServerError,
					"something went wrong",
					nil,
				)
			}
		}
	}
}
