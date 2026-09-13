package handler

import (
	"context"

	"github.com/study/jobboard/auth-service/internal/authtoken"
)

type ctxKey string

const claimsKey ctxKey = "auth_claims"

// ContextWithClaims stores JWT claims in request context.
func ContextWithClaims(ctx context.Context, claims *authtoken.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext reads JWT claims from context.
func ClaimsFromContext(ctx context.Context) *authtoken.Claims {
	claims, _ := ctx.Value(claimsKey).(*authtoken.Claims)
	return claims
}
