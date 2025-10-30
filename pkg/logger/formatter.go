package logger

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	None      = ""
	NormalTag = "NORMAL"
	FatalTag  = "FATAL"
)
const (
	RequestIDKey    = "requestID"
	RequestRouteKey = "requestRout"
	CustomTagKey    = "customTag"
)

// Formatter - logrus formatter, implements logrus.Formatter
type Formatter struct {
	FieldsOrder     []string // default: fields sorted alphabetically
	TimestampFormat string   // default: time.StampMilli = "Jan _2 15:04:05.000"
	HideKeys        bool     // show [fieldValue] instead of [fieldKey:fieldValue]
	NoColors        bool     // disable colors
	NoFieldsColors  bool     // color only level, default is level + fields
	ShowFullLevel   bool     // true to show full level [WARNING] instead [WARN]
	TrimMessages    bool     // true to trim whitespace on messages
}

// Format an log entry
// [时间]|[级别]|[类.函数]|[行数]|[终端请求编号]|自定义消息
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	baseF := "[%s]|[%s]|[%s]|[%d]|[%s]|%s\n"
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}
	t := entry.Time.Format(timestampFormat)
	level := strings.ToUpper(entry.Level.String())
	requestID := ""
	if v, ok := entry.Data[RequestIDKey]; ok {
		requestID = v.(string)
	}
	_, file := path.Split(entry.Caller.File)

	// output buffer
	b := &bytes.Buffer{}
	_, _ = fmt.Fprintf(b, baseF, t, level, file, entry.Caller.Line, requestID, entry.Message)
	return b.Bytes(), nil
}
