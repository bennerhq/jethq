package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultLevel = zerolog.ErrorLevel
	envLogLevel  = "JETHQ_LOG_LEVEL"
	// legacyEnvLogLevel is the name used before the jethq rename.
	legacyEnvLogLevel = "JETKVM_DESKTOP_LOG_LEVEL"
)

var (
	mu   sync.RWMutex
	base = newLogger(defaultLevel)
)

func logLevelFromEnv() (name, value string) {
	for _, name := range []string{envLogLevel, legacyEnvLogLevel} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return name, value
		}
	}
	return envLogLevel, ""
}

func Configure(levelText string) error {
	level := defaultLevel
	if value := strings.TrimSpace(levelText); value != "" {
		parsed, err := zerolog.ParseLevel(strings.ToLower(value))
		if err != nil {
			return fmt.Errorf("invalid log level %q: %w", value, err)
		}
		level = parsed
	} else if name, value := logLevelFromEnv(); value != "" {
		parsed, err := zerolog.ParseLevel(strings.ToLower(value))
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", name, value, err)
		}
		level = parsed
	}

	mu.Lock()
	base = newLogger(level)
	mu.Unlock()
	return nil
}

func Subsystem(name string) zerolog.Logger {
	mu.RLock()
	logger := base.With().Str("component", name).Logger()
	mu.RUnlock()
	return logger
}

func newLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(writer).Level(level).With().Timestamp().Logger()
}
