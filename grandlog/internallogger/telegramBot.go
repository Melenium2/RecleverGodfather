package internallogger

import "context"

type TelegramLogger struct {

}

func NewTelegramLogger() InternalLogger {
	return &TelegramLogger{}
}

func (t TelegramLogger) sendlog(ctx context.Context, obj interface{}) {
	panic("implement me")
}
