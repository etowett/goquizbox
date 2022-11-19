package ctxhelper

import (
	"context"
)

const ipAddressKey = "ipAddress"

func IPAddress(ctx context.Context) string {
	existing := ctx.Value(ipAddressKey)
	if existing == nil {
		return ""
	}

	return existing.(string)
}

func WithIpAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, ipAddressKey, ipAddress)
}
