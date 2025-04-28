package utils

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
	))
}
