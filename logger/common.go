package logger

import (
	"github.com/sedmess/go-ctx/u"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

var lDebug = log.New(os.Stdout, "DEBUG ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds|log.Lshortfile)
var lInfo = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds|log.Lshortfile)
var lError = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds|log.Lshortfile)
var lFatal = log.New(os.Stderr, "FATAL ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds|log.Lshortfile)

//goland:noinspection GoUnusedConst
const (
	DEBUG = iota
	INFO
	ERROR
	FATAL
)

const defaultLogLevel = INFO

var mu = sync.Mutex{}
var logLevel = int32(defaultLogLevel)

func init() {
	Init(defaultLogLevel)
}

func Init(level int) {
	mu.Lock()

	if DEBUG < level {
		lDebug.SetOutput(io.Discard)
	} else {
		lDebug.SetOutput(os.Stdout)
	}
	if INFO < level {
		lInfo.SetOutput(io.Discard)
	} else {
		lInfo.SetOutput(os.Stdout)
	}
	lError.SetOutput(os.Stderr)
	lFatal.SetOutput(os.Stderr)

	atomic.StoreInt32(&logLevel, int32(level))

	mu.Unlock()
}

func SetWriter(w io.Writer) {
	mu.Lock()

	level := atomic.LoadInt32(&logLevel)

	if DEBUG < level {
		lDebug.SetOutput(io.Discard)
	} else {
		lDebug.SetOutput(u.NewSpyWriter(w, os.Stdout))
	}
	if INFO < level {
		lInfo.SetOutput(io.Discard)
	} else {
		lInfo.SetOutput(u.NewSpyWriter(w, os.Stdout))
	}
	lError.SetOutput(u.NewSpyWriter(w, os.Stderr))
	lFatal.SetOutput(u.NewSpyWriter(w, os.Stderr))

	mu.Unlock()
}

func GetLogger(level int) *log.Logger {
	switch level {
	case DEBUG:
		return lDebug
	case INFO:
		return lInfo
	case ERROR:
		return lError
	case FATAL:
		return lFatal
	default:
		panic("unknown log level: " + strconv.Itoa(level))
	}
}

func LogLevel() int {
	return int(atomic.LoadInt32(&logLevel))
}

func DebugLazy(tag string, dataProvider func() []any) {
	if LogLevel() <= DEBUG {
		Debug(tag, dataProvider()...)
	}
}

func Debug(tag string, data ...any) {
	lDebug.Println(withTags(tag, data)...)
}

func InfoLazy(tag string, dataProvider func() []any) {
	if LogLevel() <= INFO {
		Info(tag, dataProvider()...)
	}
}

func Info(tag string, data ...any) {
	lInfo.Println(withTags(tag, data)...)
}

func Error(tag string, data ...any) {
	lError.Println(withTags(tag, data)...)
}

func Fatal(tag string, data ...any) {
	lFatal.Fatalln(withTags(tag, data)...)
}

func withTags(tag string, data []any) []any {
	if data == nil || len(data) == 0 {
		data = make([]any, 1)
	} else {
		data = append(data, nil)
		copy(data[1:], data)
	}
	data[0] = "[" + tag + "]"
	return data
}
