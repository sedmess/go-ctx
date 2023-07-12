package logger

type Logger interface {
	Debug(msg ...any)
	Info(msg ...any)
	Error(msg ...any)
	Fatal(msg ...any)
}

type logger struct {
	tag string
}

type Loggable interface {
	Name() string
}

func New(loggable Loggable) Logger {
	return NewWithTag(loggable.Name())
}

func NewWithTag(tag string) Logger {
	return &logger{tag: tag}
}

func (instance *logger) Debug(msg ...any) {
	Debug(instance.tag, msg...)
}

func (instance *logger) DebugLazy(dataProvider func() []any) {
	DebugLazy(instance.tag, dataProvider)
}

func (instance *logger) Info(msg ...any) {
	Info(instance.tag, msg...)
}

func (instance *logger) InfoLazy(dataProvider func() []any) {
	InfoLazy(instance.tag, dataProvider)
}

func (instance *logger) Error(msg ...any) {
	Error(instance.tag, msg...)
}

func (instance *logger) Fatal(msg ...any) {
	Fatal(instance.tag, msg...)
}
