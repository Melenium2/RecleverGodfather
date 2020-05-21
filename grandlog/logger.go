package grandlog

import (
	"RecleverGodfather/grandlog/internallogger"
	"RecleverGodfather/grandlog/loggerepo"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"regexp"
	"strings"
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
	logger         log.Logger
	internalLogger internallogger.InternalLogger
}

func NewGrandLogger(db loggerepo.Logs, logger log.Logger, internalLogger internallogger.InternalLogger) GrandLogger {
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
	defer gl.internalLogger.Sendlog(0, fmt.Sprintf("%v [%s] %s", time.Now().String(), msgType, log))

	return gl.repoLogger.Savelog(ctx, msgType, strings.TrimSpace(log))
}

func (gl *GrandLog) LogObject(ctx context.Context, log *loggerepo.SingleLog) error {
	defer gl.internalLogger.Sendlog(0, fmt.Sprintf("%v [%s] %s", time.Now().String(),log.MessageType, log.Log))

	return gl.repoLogger.Savelog(ctx, log.MessageType, log.Log)
}
