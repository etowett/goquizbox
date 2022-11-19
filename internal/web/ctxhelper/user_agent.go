package ctxhelper

import (
	"context"
)

const userAgentKey = "userAgent"

func UserAgent(ctx context.Context) string {
	existing := ctx.Value(userAgentKey)
	if existing == nil {
		return ""
	}

	return existing.(string)
}

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}
