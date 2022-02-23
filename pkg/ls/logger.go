package ls

import (
	"encoding/json"
	"log"
	"runtime/debug"
	"time"
)

type Logger interface {
	Debug(map[string]interface{})
	Info(map[string]interface{})
	Error(map[string]interface{})
}

type NopLogger struct{}

func (l *NopLogger) Debug(map[string]interface{}) {}
func (l *NopLogger) Info(map[string]interface{})  {}
func (l *NopLogger) Error(map[string]interface{}) {}

type LogLevel int8

const (
	LogLevelInfo LogLevel = iota
	LogLevelError
	LogLevelDebug
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelError:
		return "ERROR"
	default:
		return ""
	}
}

type DefaultLogger struct {
	minLevel LogLevel
}

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

func (l DefaultLogger) Info(properties map[string]interface{}) {
	l.print(LogLevelInfo, properties)
}

func (l DefaultLogger) Debug(properties map[string]interface{}) {
	print(LogLevelDebug, properties)
}

func (l DefaultLogger) Error(properties map[string]interface{}) {
	print(LogLevelError, properties)
}

func (l DefaultLogger) print(level LogLevel, properties map[string]interface{}) {
	aux := struct {
		Level      string                 `json:"level"`
		Time       string                 `json:"time"`
		Properties map[string]interface{} `json:"properties,omitempty"`
		Trace      string                 `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Properties: properties,
	}

	if level >= LogLevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LogLevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	log.Println(string(line))
}
