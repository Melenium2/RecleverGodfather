package internallogger

type UpdateChannel <-chan *update

func (up UpdateChannel) Clear() {
	for len(up) != 0 {
		<-up
	}
}

type UpdateConfig struct {
	Offset  int
	Limit   int
	Timeout int
}

func GenerateUpdateConfig(offset int) *UpdateConfig {
	return &UpdateConfig{
		Offset:  offset,
		Limit:   0,
		Timeout: 30,
	}
}

type response struct {
	Ok     bool      `json:"ok"`
	Result []*update `json:"result"`
}

type update struct {
	UpdateId int     `json:"update_id"`
	Message  message `json:"message"`
}

type message struct {
	MessageId int    `json:"message_id"`
	From      *user  `json:"from"`
	Chat      *chat  `json:"chat"`
	Date      int64  `json:"date"`
	Text      string `json:"text"`
}

type user struct {
	Id        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Language  string `json:"language_code"`
}

type chat struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}
