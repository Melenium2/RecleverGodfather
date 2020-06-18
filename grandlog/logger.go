package grandlog

import (
	"RecleverGodfather/grandlog/internallogger"
	"RecleverGodfather/grandlog/loggerepo"
	"context"
	"fmt"
	murlog "github.com/Melenium2/Murlog"
	"regexp"
	"time"
)

type GrandLogger interface {
	Log(...interface{}) error
	Logs(context.Context, ...interface{}) error
	LogObject(context.Context, *loggerepo.SingleLog) error
}

var (
	reMsg = regexp.MustCompile(`\[\w+]`)
	reRedundant = regexp.MustCompile(`[\[\]]`)
)

type GrandLog struct {
	repoLogger     loggerepo.Logs
	logger         murlog.Logger
	internalLogger internallogger.InternalLogger
}

func NewGrandLogger(db loggerepo.Logs, logger murlog.Logger, internalLogger internallogger.InternalLogger) GrandLogger {
	return &GrandLog{
		repoLogger:     db,
		logger:         logger,
		internalLogger: internalLogger,
	}
}

func (gl *GrandLog) Log(keyvals ...interface{}) error {
	return gl.logger.Log(keyvals...)
}

func (gl *GrandLog) Logs(ctx context.Context, keyvals ...interface{}) error {
	msgType := reMsg.FindString(keyvals[0].(string))
	msgType = reRedundant.ReplaceAllString(msgType, "")

	var log string
	{
		for i, str := range keyvals[1:] {
			sep := " "
			if i % 2 == 0 {
				sep = "="
			}
			log += str.(string) + sep
		}
	}

	gl.LogObject(ctx, &loggerepo.SingleLog{
		MessageType: msgType,
		Log:         log,
	})

	return gl.Log(keyvals...)
}

func (gl *GrandLog) LogObject(ctx context.Context, log *loggerepo.SingleLog) error {
	if log.MessageType == "Error" {
		defer gl.internalLogger.Sendlog(0, fmt.Sprintf("%v [%s] %v", timestamp(), log.MessageType, log.Log))
	}
	return gl.repoLogger.Savelog(ctx, log.MessageType, log.Log)
}

func timestamp() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}