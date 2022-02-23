package ls

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"runtime/debug"
	"time"
)

type Logger interface {
	Debug(map[string]interface{})
	Info(map[string]interface{})
	Error(map[string]interface{})
}

type NopLogger struct {
	*log.Logger
}

func (l *NopLogger) Debug(map[string]interface{}) {}
func (l *NopLogger) Info(map[string]interface{})  {}
func (l *NopLogger) Error(map[string]interface{}) {}

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelDebug
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type DefaultLogger struct {
	log      *log.Logger
	minLevel Level
}

func New(log *log.Logger, minLevel Level) *DefaultLogger {
	return &DefaultLogger{
		log:      log,
		minLevel: minLevel,
	}
}

func (l DefaultLogger) Info(properties map[string]interface{}) {
	l.print(LevelInfo, properties)
}

func (l DefaultLogger) Debug(properties map[string]interface{}) {
	print(LevelDebug, properties)
}

func (l DefaultLogger) Error(properties map[string]interface{}) {
	print(LevelError, properties)
}

func (l DefaultLogger) Fatal(properties map[string]interface{}) {
	print(LevelFatal, properties)
	os.Exit(1)
}

func (l DefaultLogger) print(level Level, properties map[string]interface{}) (int, error) {
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

	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	l.log.Println(line)
	return bytes.Count(line, []byte{}), nil
}
