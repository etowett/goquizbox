package logger

import (
	"context"

	"goquizbox/internal/web/ctxhelper"
)

func unwrapContext(ctx context.Context) []Field {
	fields := []Field{}

	requestId := ctxhelper.RequestId(ctx)
	if requestId != "" {
		fields = append(
			fields,
			String("request_id", requestId),
		)
	}

	return fields
}
