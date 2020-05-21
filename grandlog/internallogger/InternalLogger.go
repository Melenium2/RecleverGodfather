package internallogger

type InternalLogger interface {
	Sendlog(chatId int, message string) error
}
