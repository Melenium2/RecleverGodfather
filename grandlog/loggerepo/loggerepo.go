package loggerepo

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"time"
)

type SingleLog struct {
	MessageType string    `db:"message_type" json:"message_type"`
	Times       time.Time `db:"times" json:"times"`
	Log         string    `db:"log" json:"log"`
}

type Logs interface {
	Savelog(context.Context, string, string) error
}

type ClickhouseLogger struct {
	logger log.Logger
	db     *sqlx.DB
}

func NewClickhouseLogger(db *sqlx.DB, log log.Logger) Logs {
	return &ClickhouseLogger{
		db:     db,
		logger: log,
	}
}

func (c *ClickhouseLogger) Savelog(ctx context.Context, messageType, log string) error {
	if _, err := c.db.ExecContext(
		ctx,
		`insert into logs (message_type, times, log) values (?, ?, ?)`,
		messageType, time.Now().UTC().Unix(), log,
	); err != nil {
		c.logger.Log("[Error]", err)
		return err
	}
	return nil
}
