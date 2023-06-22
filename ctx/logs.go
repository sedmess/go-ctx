package ctx

import (
	"github.com/sedmess/go-ctx/logger"
)

const debugLoggingParam = "DEBUG_LOG_LEVEL"

func init() {
	if GetEnv(debugLoggingParam).AsBoolDefault(false) {
		logger.Init(logger.DEBUG)
	}
}

// IsDebugLogEnabled
// Deprecated: use logger.Configure
func IsDebugLogEnabled() bool {
	return logger.LogLevel() <= logger.DEBUG
}

// LogDebug
// Deprecated: use logger.Debug
func LogDebug(tag string, data ...any) {
	logger.Debug(tag, data...)
}

// LogInfo
// Deprecated: use logger.Debug
func LogInfo(tag string, data ...any) {
	logger.Info(tag, data...)
}

// LogError
// Deprecated: use logger.Debug
func LogError(tag string, data ...any) {
	logger.Error(tag, data...)
}

// LogFatal
// Deprecated: use logger.Debug
func LogFatal(tag string, data ...any) {
	logger.Fatal(tag, data...)
}
