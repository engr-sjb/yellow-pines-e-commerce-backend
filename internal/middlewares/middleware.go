package middlewares

import "github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"

type tokenManager interface {
	ValidateAccessToken(tokenStr string) (isValid bool, claims *auth.TokenClaims, err error)
}

type middleware struct {
	jwtManager tokenManager
}

func NewMiddleware(tokenManager tokenManager) *middleware {
	return &middleware{
		jwtManager: tokenManager,
	}
}
