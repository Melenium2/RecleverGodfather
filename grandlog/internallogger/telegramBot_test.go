package internallogger

import (
	"fmt"
	"testing"
)

// Работает некорректно вылетает ошибка при закрытии канала
func TestNewTelegramLogger(t *testing.T) {
	bot := NewTelegramLogger("1293039613:AAER81Qqklo9JZQa3kt2iHrKBA9ptPpJ8IY", 708015155)
	up := bot.ServeUpdates(GenerateUpdateConfig(0))
	for  {
		update, ok := <-up
		println(ok)
		if !ok {
			break
		}
		println(fmt.Sprintf("%v", update.Message.Text))
		bot.Sendlog(update.Message.Chat.Id, "Privet")
		bot.CloseUpdates()
	}
}
