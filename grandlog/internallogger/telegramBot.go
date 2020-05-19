package internallogger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const ApiEndpoint = "https://api.telegram.org/bot%s/%s"

type TelegramLogger struct {
	apiToken     string
	shutdownChan chan interface{}
	client       *http.Client
}

func NewTelegramLogger(token string) *TelegramLogger {
	return &TelegramLogger{
		apiToken: token,
		shutdownChan: make(chan interface{}),
		client: http.DefaultClient,
	}
}

func (t *TelegramLogger) Sendlog(ctx context.Context, obj interface{}) {
	panic("implement me")
}

func (t *TelegramLogger) GetUpdates(config *UpdateConfig) (*response, error) {
	v := url.Values{}
	if config.Offset != 0 {
		v.Add("offset", strconv.Itoa(config.Offset))
	}
	if config.Limit > 0 {
		v.Add("limit", strconv.Itoa(config.Limit))
	}
	if config.Timeout > 0 {
		v.Add("limit", strconv.Itoa(config.Timeout))
	}

	b, err := t.doRequest("getUpdates", v)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	resp := &response{}
	if err := json.NewDecoder(b).Decode(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (t *TelegramLogger) ServeUpdates(config *UpdateConfig) (UpdateChannel, error) {
	up := make(chan *update)

	go func() {
		for {
			select {
			case <-t.shutdownChan:
				close(up)
				return
			default:
			}

			updates, err := t.GetUpdates(config)
			if err != nil {
				println(err.Error())
				time.Sleep(time.Second * 5)
				continue
			}

			for _, update := range updates.Result {
				println(update.UpdateId)
				if update.UpdateId >= config.Offset {
					config.Offset = update.UpdateId + 1
					up <- update
				}
			}
		}
	}()

	return up, nil
}

func (t *TelegramLogger) CloseUpdates() {
	close(t.shutdownChan)
}

func (t *TelegramLogger) doRequest(endpoint string, params url.Values) (io.ReadCloser, error) {
	url := fmt.Sprintf(ApiEndpoint, t.apiToken, endpoint)
	resp, err := t.client.PostForm(url, params)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
