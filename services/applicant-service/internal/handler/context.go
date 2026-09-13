package handler

import (
	"context"

	"github.com/study/jobboard/applicant-service/internal/authtoken"
)

type ctxKey string

const claimsKey ctxKey = "auth_claims"

func ContextWithClaims(ctx context.Context, claims *authtoken.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) *authtoken.Claims {
	claims, _ := ctx.Value(claimsKey).(*authtoken.Claims)
	return claims
}
