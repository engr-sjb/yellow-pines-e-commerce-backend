package middlewares

import (
	"context"
	"net/http"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type contextKey struct{}

var EntityKey contextKey = contextKey{}

func (mw *middleware) AuthWithContext(h handlerutils.APIHandler, authEntityType string) handlerutils.APIHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		accessToken, err := r.Cookie("accessToken")
		if err != nil {
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrNoAccessTokenCookie.Error(),
				nil,
			)
		}

		isValid, claims, err := mw.jwtManager.ValidateAccessToken(accessToken.Value)
		if err != nil {
			return err
		}

		if !isValid {
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrUnauthorized.Error(),
				nil,
			)
		}

		if claims.EntityType != authEntityType {
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrUnauthorizedAccess.Error(),
				nil,
			)
		}

		// create context
		ctx := r.Context()
		ctx = context.WithValue(
			ctx,
			EntityKey,
			claims.EntityID,
		)
		r = r.WithContext(ctx)

		return h(w, r)
	}
}

func GetEntityIDFromContextKey(ctx context.Context) uuid.UUID {
	entityIDStr, ok := ctx.Value(EntityKey).(string)
	if !ok {
		return uuid.Nil
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return uuid.Nil
	}

	return entityID
}
