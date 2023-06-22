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
	return &logger{tag: loggable.Name()}
}

func (instance *logger) Debug(msg ...any) {
	Debug(instance.tag, msg...)
}

func (instance *logger) Info(msg ...any) {
	Info(instance.tag, msg...)
}

func (instance *logger) Error(msg ...any) {
	Error(instance.tag, msg...)
}

func (instance *logger) Fatal(msg ...any) {
	Fatal(instance.tag, msg...)
}
