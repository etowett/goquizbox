package ctxhelper

import (
	"context"

	"goquizbox/internal/entities"
)

const tokenInfoKey = "tokenInfo"

func TokenInfo(ctx context.Context) *entities.TokenInfo {
	existing := ctx.Value(tokenInfoKey)
	if existing == nil {
		return &entities.TokenInfo{}
	}

	tokenInfo, ok := existing.(*entities.TokenInfo)
	if !ok {
		return &entities.TokenInfo{}
	}

	return tokenInfo
}

func WithTokenInfo(ctx context.Context, tokenInfo *entities.TokenInfo) context.Context {
	return context.WithValue(ctx, tokenInfoKey, tokenInfo)
}

func UserID(ctx context.Context) int64 {
	return TokenInfo(ctx).UserID
}
