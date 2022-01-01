package ctx

import "log"

const debugLoggingParam = "DEBUG_LOG_LEVEL"

var debugEnabled = false

func init() {
	debugEnabled = GetEnv(debugLoggingParam).AsBoolDefault(false)
}

//goland:noinspection GoUnusedExportedFunction
func IsDebugLogEnabled() bool {
	return debugEnabled
}

func LogDebug(tag string, data ...interface{}) {
	if debugEnabled {
		log.Println(withTags(tag, "DEBUG", data)...)
	}
}

func LogInfo(tag string, data ...interface{}) {
	log.Println(withTags(tag, "INFO", data)...)
}

func LogError(tag string, data ...interface{}) {
	log.Println(withTags(tag, "ERROR", data)...)
}

func LogFatal(tag string, data ...interface{}) {
	log.Fatalln(withTags(tag, "FATAL", data)...)
}

func withTags(tag string, level string, data []interface{}) []interface{} {
	if data == nil || len(data) == 0 {
		data = make([]interface{}, 2)
	} else {
		data = append(data, nil, nil)
		copy(data[2:], data)
	}
	data[0] = "[" + tag + "]"
	data[1] = level + ":"
	return data
}
