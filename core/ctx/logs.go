package ctx

import (
	"github.com/sedmess/go-ctx/logger"
	"strings"
)

const debugLoggingParam = "DEBUG_LOG_LEVEL"
const logLevelParam = "LOG_LEVEL"

func init() {
	if GetEnv(debugLoggingParam).AsBoolDefault(false) {
		logger.Init(logger.DEBUG)
	} else {
		switch strings.ToUpper(GetEnv(logLevelParam).AsStringDefault("INFO")) {
		case "DEBUG":
			logger.Init(logger.DEBUG)
		case "INFO":
			logger.Init(logger.INFO)
		case "ERROR":
			logger.Init(logger.ERROR)
		default:
			logger.Init(logger.INFO)
		}
	}
}
