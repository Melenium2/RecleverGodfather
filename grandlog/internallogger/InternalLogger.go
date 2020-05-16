package internallogger

import "context"

type InternalLogger interface {
	sendlog(ctx context.Context, obj interface{})
}
