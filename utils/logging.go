package utils

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"io"
	"os"
	"sync"
	"time"
)

// Gorm logger -------------------------
type GormLogger struct{}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l GormLogger) Info(ctx context.Context, a string, b ...interface{}) {
	log.FromContext(ctx).Infof(a, b)
}

func (l GormLogger) Warn(ctx context.Context, a string, b ...interface{}) {
	log.FromContext(ctx).Warnf(a, b)
}

func (l GormLogger) Error(ctx context.Context, a string, b ...interface{}) {
	log.FromContext(ctx).Errorf(a, b)
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	logger_ := log.FromContext(ctx)
	logger_ = logger_.
		WithError(err).
		WithFields(log.Fields{
			"line":         utils.FileWithLineNum(),
			"duration":     fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
			"rowsAffected": rows,
			"query":        sql,
		})

	if err != nil {
		logger_.Error("Database query error")
	}

	if elapsed > 200*time.Millisecond {
		logger_.Warnf("Database query took more than %dms", 200)
	}

	logger_.Debug("Database query")
}

// Terminal logger -------------------------
var colorValue = color.New(color.FgCyan)
var colorField = color.New(color.FgBlue)

var Colors = [...]*color.Color{
	log.DebugLevel: color.New(color.FgHiBlue),
	log.InfoLevel:  color.New(color.FgGreen),
	log.WarnLevel:  color.New(color.FgYellow),
	log.ErrorLevel: color.New(color.FgRed),
	log.FatalLevel: color.New(color.BgRed + color.FgWhite),
}

type TerminalLogger struct {
	mu     sync.Mutex
	Writer io.Writer
}

func NewTerminalLogger(w io.Writer) *TerminalLogger {
	if f, ok := w.(*os.File); ok {
		return &TerminalLogger{
			Writer: colorable.NewColorable(f),
		}
	}

	return &TerminalLogger{
		Writer: w,
	}
}

// HandleLog implements log.Handler.
func (h *TerminalLogger) HandleLog(e *log.Entry) error {
	color_ := Colors[e.Level]
	names := e.Fields.Names()

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := color_.Fprintf(h.Writer, "[%s] %-30s", e.Timestamp.Format("15:04:05"), color_.Sprint(e.Message))
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == "source" {
			continue
		}
		_, err = fmt.Fprintf(h.Writer, " %s=%v", colorField.Sprint(name), colorValue.Sprint(e.Fields.Get(name)))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(h.Writer)
	if err != nil {
		return err
	}

	return nil
}
