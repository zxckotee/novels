package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New создает новый экземпляр логгера
func New(isDev bool) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if isDev {
		// В development режиме используем красивый вывод
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
		return zerolog.New(output).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	// В production режиме используем JSON формат
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
}
