package console

import (
	"errors"
	"os"

	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/logs"
	"github.com/cgalvisleon/elvis/strs"
)

func NewError(message string) error {
	err := errors.New(message)

	return err
}

func NewErrorF(format string, args ...any) error {
	message := strs.Format(format, args...)
	err := NewError(message)

	return err
}

func LogK(kind string, args ...any) error {
	event.Action("logs", map[string]interface{}{
		"kind": kind,
		"args": args,
	})

	logs.Log(kind, args...)

	return nil
}

func LogKF(kind string, format string, args ...any) error {
	message := strs.Format(format, args...)
	return LogK(kind, message)
}

func Log(args ...any) error {
	return LogK("Log", args...)
}

func LogF(format string, args ...any) error {
	message := strs.Format(format, args...)
	return Log(message)
}

func Debug(args ...any) error {
	logs.Debug(args...)
	return nil
}

func DebugF(format string, args ...any) error {
	message := strs.Format(format, args...)
	return Debug(message)
}

func Print(args ...any) error {
	message := ""
	for i, arg := range args {
		if i == 0 {
			message = strs.Format("%v", arg)
		} else {
			message = strs.Format("%s, %v", message, arg)
		}
	}
	return Log(message)
}

func Info(args ...any) error {
	logs.Info(args...)
	return nil
}

func InfoF(format string, args ...any) error {
	message := strs.Format(format, args...)
	return Info(message)
}

func Alert(message string) error {
	return logs.Alertm(message)
}

func AlertF(format string, args ...any) error {
	return logs.Alertf(format, args...)
}

func Error(err error) error {
	traces, err := logs.Traces("Error", "Red", err)

	event.Action("logs", map[string]interface{}{
		"kind": "ERROR",
		"args": traces,
	})

	return err
}

func ErrorM(message string) error {
	err := NewError(message)
	return Error(err)
}

func ErrorF(format string, args ...any) error {
	message := strs.Format(format, args...)
	err := NewError(message)
	return Error(err)
}

func Fatal(v ...any) {
	logs.Fatal(v...)
	os.Exit(1)
}

func FatalF(format string, args ...any) {
	message := strs.Format(format, args...)
	Fatal(message)
}

func Panic(err error) error {
	traces, err := logs.Traces("Panic", "Red", err)

	event.Action("logs", map[string]interface{}{
		"kind": "ERROR",
		"args": traces,
	})

	os.Exit(1)

	return err
}

func PanicM(message string) error {
	err := ErrorM(message)
	Panic(err)
	return err
}

func PanicF(format string, args ...any) error {
	err := ErrorF(format, args...)
	Panic(err)
	return err
}

func Ping() {
	Log("PING")
}

func Pong() {
	Log("PONG")
}
