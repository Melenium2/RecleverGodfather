package internallogger

import (
	"fmt"
	"testing"
)

func TestNewTelegramLogger(t *testing.T) {
	bot := NewTelegramLogger("1293039613:AAER81Qqklo9JZQa3kt2iHrKBA9ptPpJ8IY")
	up, err := bot.ServeUpdates(GenerateUpdateConfig(0))
	if err != nil {
		t.Fatal(err)
	}
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
