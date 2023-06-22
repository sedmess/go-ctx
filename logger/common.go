package logger

import (
	"github.com/sedmess/go-ctx/u"
	"io"
	"log"
	"os"
	"sync"
)

var lDebug = log.New(os.Stdout, "DEBUG", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds)
var lInfo = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds)
var lError = log.New(os.Stderr, "ERROR", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds)
var lFatal = log.New(os.Stderr, "FATAL", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lmicroseconds)

//goland:noinspection GoUnusedConst
const (
	DEBUG = iota
	INFO
	ERROR
)

const defaultLogLevel = INFO

var mu = sync.Mutex{}
var logLevel = defaultLogLevel

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

	logLevel = level

	mu.Unlock()
}

func SetWriter(w io.Writer) {
	mu.Lock()

	if DEBUG < logLevel {
		lDebug.SetOutput(io.Discard)
	} else {
		lDebug.SetOutput(u.NewSpyWriter(w, os.Stdout))
	}
	if INFO < logLevel {
		lInfo.SetOutput(io.Discard)
	} else {
		lInfo.SetOutput(u.NewSpyWriter(w, os.Stdout))
	}
	lError.SetOutput(u.NewSpyWriter(w, os.Stderr))
	lFatal.SetOutput(u.NewSpyWriter(w, os.Stderr))

	mu.Unlock()
}

func LogLevel() int {
	mu.Lock()
	r := logLevel
	mu.Unlock()
	return r
}

func Debug(tag string, data ...any) {
	lDebug.Println(withTags(tag, data)...)
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
	data[0] = " [" + tag + "]"
	return data
}
