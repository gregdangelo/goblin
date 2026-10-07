package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewJSONLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "msg" {
				a.Key = "message"
			} else if a.Key == "time" {
				a.Key = "timestamp"
			} else if a.Key == "level" {
				a.Value = slog.StringValue(strings.ToLower(a.Value.String()))
			}
			return a
		},
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
