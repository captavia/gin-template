package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

func NewFormatter() logrus.Formatter {
	return new(CustomFormatter)
}

type CustomFormatter struct{}

func (t *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	_, _ = fmt.Fprintf(b, "[%s]", entry.Time.Format("2006-01-02 15:04:05"))

	_, _ = fmt.Fprintf(b, "\u001B[%dm[%s]\x1b[0m", levelColor, entry.Level)

	if entry.HasCaller() {
		_, _ = fmt.Fprintf(b, "[%s:%d]", path.Base(entry.Caller.File), entry.Caller.Line)
	}
	if len(entry.Data) != 0 {
		for f, v := range entry.Data {
			if v == "" {
				_, _ = fmt.Fprintf(b, "[%s]", f)
			} else {
				_, _ = fmt.Fprintf(b, "[%s:%s]", f, v)
			}
		}
	}
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}
