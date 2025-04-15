package logger

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
)

const TagKey = "tag"

type slogWrapperStruct struct {
	mu      sync.RWMutex
	handler slog.Handler
}

var slogWrapper = slogWrapperStruct{
	handler: slog.NewTextHandler(os.Stdout, nil),
}

func SetSlogHandler(handler slog.Handler) {
	slogWrapper.mu.Lock()
	defer slogWrapper.mu.Unlock()

	if handler == nil {
		log.Fatal("handler cannot be nil")
	}

	slogWrapper.handler = handler
}

func CreateSlogFor(serviceName string) *slog.Logger {
	slogWrapper.mu.RLock()
	defer slogWrapper.mu.RUnlock()

	return slog.New(slogWrapper.handler).With(TagKey, serviceName)
}

type Logger interface {
	Debug(msg ...any)
	Info(msg ...any)
	Error(msg ...any)
	Fatal(msg ...any)
}

type logger struct {
	l *slog.Logger
}

func New(serviceName string) Logger {
	return &logger{l: CreateSlogFor(serviceName)}
}

func (instance *logger) Debug(msg ...any) {
	writeLog(instance.l.Handler(), slog.LevelDebug, msg...)
}

func (instance *logger) Info(msg ...any) {
	writeLog(instance.l.Handler(), slog.LevelInfo, msg...)
}

func (instance *logger) Warn(msg ...any) {
	writeLog(instance.l.Handler(), slog.LevelWarn, msg...)
}

func (instance *logger) Error(msg ...any) {
	writeLog(instance.l.Handler(), slog.LevelError, msg...)
}

func (instance *logger) Fatal(msg ...any) {
	writeLog(instance.l.Handler(), slog.LevelError, msg...)
	log.Fatalln("fatal error occurred")
}

func Debug(tag string, msg ...any) {
	writeLog(slog.Default().With(slog.String(TagKey, tag)).Handler(), slog.LevelDebug, msg...)
}

func Info(tag string, msg ...any) {
	writeLog(slog.Default().With(slog.String(TagKey, tag)).Handler(), slog.LevelInfo, msg...)
}

func Warn(tag string, msg ...any) {
	writeLog(slog.Default().With(slog.String(TagKey, tag)).Handler(), slog.LevelWarn, msg...)
}

func Error(tag string, msg ...any) {
	writeLog(slog.Default().With(slog.String(TagKey, tag)).Handler(), slog.LevelError, msg...)
}

func Fatal(tag string, msg ...any) {
	writeLog(slog.Default().With(slog.String(TagKey, tag)).Handler(), slog.LevelError, msg...)
	log.Fatalln("fatal error occurred")
}

func writeLog(handler slog.Handler, level slog.Level, msg ...any) {
	ctx := context.Background()
	if !handler.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, fmt.Sprintln(msg...), pc)
	_ = handler.Handle(ctx, r)
}
