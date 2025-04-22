package ctx

import (
	"context"
	"github.com/sedmess/go-ctx/ctx/appinfo"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

const (
	slogHandlerParam       = "SLOG_HANDLER"
	slogAddSourceParam     = "SLOG_ADD_SOURCE"
	slogAddCommonTagsParam = "SLOG_ADD_COMMON_TAGS"
	slogLevelParam         = "SLOG_LEVEL"
)

const (
	slogHandlerLegacy = "legacy"
	slogHandlerText   = "text"
	slogHandlerJson   = "json"
)

const (
	slogLevelDebug = "debug"
	slogLevelInfo  = "info"
	slogLevelWarn  = "warn"
	slogLevelError = "error"
)

var slogHandlerSet = atomic.Bool{}

type slogLegacyHandler struct {
	level  slog.Level
	tag    string
	lDebug *log.Logger
	lInfo  *log.Logger
	lWarn  *log.Logger
	lError *log.Logger
	lFatal *log.Logger
}

func NewSlogLegacyHandler(w io.Writer, level slog.Level, addSource bool) slog.Handler {
	flags := log.Ldate | log.Ltime | log.Lmsgprefix | log.Lmicroseconds
	if addSource {
		flags |= log.Lshortfile
	}
	handler := &slogLegacyHandler{
		level:  level,
		lError: log.New(w, "ERROR ", flags),
		lFatal: log.New(w, "FATAL ", flags),
	}
	if level <= slog.LevelWarn {
		handler.lWarn = log.New(w, "WARN ", flags)
	} else {
		handler.lWarn = log.New(io.Discard, "WARN ", flags)
	}
	if level <= slog.LevelInfo {
		handler.lInfo = log.New(w, "INFO ", flags)
	} else {
		handler.lInfo = log.New(io.Discard, "INFO ", flags)
	}
	if level <= slog.LevelDebug {
		handler.lDebug = log.New(w, "DEBUG ", flags)
	} else {
		handler.lDebug = log.New(io.Discard, "DEBUG ", flags)
	}

	return handler
}

func (s *slogLegacyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return s.level <= level
}

func (s *slogLegacyHandler) WithAttrs(attr []slog.Attr) slog.Handler {
	for i := range attr {
		if attr[i].Key == logger.TagKey {
			newHandler := s.copy()
			newHandler.tag = "[" + attr[i].Value.String() + "] "
			return newHandler
		}
	}
	// not supported
	return s
}

func (s *slogLegacyHandler) WithGroup(_ string) slog.Handler {
	// not supported
	return s
}

func (s *slogLegacyHandler) copy() *slogLegacyHandler {
	return &slogLegacyHandler{
		level:  s.level,
		lDebug: s.lDebug,
		lInfo:  s.lInfo,
		lWarn:  s.lWarn,
		lError: s.lError,
		lFatal: s.lFatal,
	}
}

func (s *slogLegacyHandler) Handle(_ context.Context, record slog.Record) error {
	switch record.Level {
	case slog.LevelDebug:
		return s.lDebug.Output(4, s.tag+record.Message)
	case slog.LevelInfo:
		return s.lInfo.Output(4, s.tag+record.Message)
	case slog.LevelWarn:
		return s.lWarn.Output(4, s.tag+record.Message)
	case slog.LevelError:
		return s.lError.Output(4, s.tag+record.Message)
	default:
		return s.lInfo.Output(4, s.tag+record.Message)
	}
}

//goland:noinspection GoBoolExpressions
func createSlogHandler(writer io.Writer) slog.Handler {
	var leveler slog.Leveler

	switch strings.ToLower(GetEnv(slogLevelParam).AsStringDefault(slogLevelInfo)) {
	case slogLevelDebug:
		leveler = slog.LevelDebug
	case slogLevelInfo:
		leveler = slog.LevelInfo
	case slogLevelWarn:
		leveler = slog.LevelWarn
	case slogLevelError:
		leveler = slog.LevelError
	default:
		log.Fatal("SLOG_LEVEL must be one of [debug, info, warn, error]")
	}

	addSource := GetEnv(slogAddSourceParam).AsBoolDefault(false)

	var handler slog.Handler
	handlerOptions := &slog.HandlerOptions{Level: leveler, AddSource: addSource}

	switch strings.ToLower(GetEnv(slogHandlerParam).AsStringDefault(slogHandlerText)) {
	case slogHandlerLegacy:
		handler = NewSlogLegacyHandler(writer, handlerOptions.Level.Level(), handlerOptions.AddSource)
	case slogHandlerText:
		handler = slog.NewTextHandler(writer, handlerOptions)
	case slogHandlerJson:
		handler = slog.NewJSONHandler(writer, handlerOptions)
	default:
		log.Fatal("SLOG_HANDLER must be one of [text, json, legacy]")
	}

	return handler
}

func prepareSlogHandler(h slog.Handler) (handler slog.Handler) {
	handler = h
	if GetEnv(slogAddCommonTagsParam).AsBoolDefault(false) {
		var attrs []slog.Attr
		if appinfo.Name != "" {
			attrs = append(attrs, slog.String("appName", appinfo.Name))
		}
		if //goland:noinspection GoBoolExpressions
		appinfo.Version != "" {
			attrs = append(attrs, slog.String("appVersion", appinfo.Version))
		}
		if //goland:noinspection GoBoolExpressions
		appinfo.BuildInfo != "" {
			attrs = append(attrs, slog.String("buildInfo", appinfo.BuildInfo))
		}
		if len(attrs) > 0 {
			handler = handler.WithAttrs(attrs)
		}
	}
	return
}

func InitSlog() {
	if slogHandlerSet.Load() {
		return
	}
	handler := prepareSlogHandler(createSlogHandler(os.Stdout))
	logger.SetSlogHandler(handler)
	slog.SetDefault(slog.New(handler))
}

func SetSlogWriter(writers ...io.Writer) {
	if len(writers) == 0 {
		log.Fatal("writers must not be empty")
	}
	w := writers[0]
	for i := 1; i < len(writers); i++ {
		w = u.NewSpyWriter(w, writers[i])
	}
	handler := prepareSlogHandler(createSlogHandler(w))

	slogHandlerSet.Store(true)

	logger.SetSlogHandler(handler)
	slog.SetDefault(slog.New(handler))
}

//goland:noinspection GoUnusedExportedFunction
func SetSlogHandler(handler slog.Handler) {
	if handler == nil {
		log.Fatal("handler cannot be nil")
	}
	handler = prepareSlogHandler(handler)

	slogHandlerSet.Store(true)

	logger.SetSlogHandler(handler)
	slog.SetDefault(slog.New(handler))
}
