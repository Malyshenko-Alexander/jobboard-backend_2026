package handler

import (
	"context"

	"github.com/study/jobboard/vacancy-service/internal/authtoken"
)

type ctxKey string

const (
	claimsKey ctxKey = "auth_claims"
	tokenKey  ctxKey = "raw_bearer_token"
)

func ContextWithClaims(ctx context.Context, claims *authtoken.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) *authtoken.Claims {
	claims, _ := ctx.Value(claimsKey).(*authtoken.Claims)
	return claims
}

func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func TokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(tokenKey).(string)
	return token
}
