package logger

type Logger interface {
	Debug(msg ...any)
	DebugLazy(dataProvider func() []any)
	Info(msg ...any)
	InfoLazy(dataProvider func() []any)
	Error(msg ...any)
	Fatal(msg ...any)
	LogDebug(msg ...any)
	LogDebugLazy(dataProvider func() []any)
	LogInfo(msg ...any)
	LogInfoLazy(dataProvider func() []any)
	LogError(msg ...any)
	LogIfError(when string, err error)
	LogFatal(msg ...any)
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
	debugInt(instance.tag, msg...)
}

func (instance *logger) DebugLazy(dataProvider func() []any) {
	debugLazyInt(instance.tag, dataProvider)
}

func (instance *logger) Info(msg ...any) {
	infoInt(instance.tag, msg...)
}

func (instance *logger) InfoLazy(dataProvider func() []any) {
	infoLazyInt(instance.tag, dataProvider)
}

func (instance *logger) Error(msg ...any) {
	errorInt(instance.tag, msg...)
}

func (instance *logger) Fatal(msg ...any) {
	fatalInt(instance.tag, msg...)
}

func (instance *logger) LogDebug(msg ...any) {
	debugInt(instance.tag, msg...)
}

func (instance *logger) LogDebugLazy(dataProvider func() []any) {
	debugLazyInt(instance.tag, dataProvider)
}

func (instance *logger) LogInfo(msg ...any) {
	infoInt(instance.tag, msg...)
}

func (instance *logger) LogInfoLazy(dataProvider func() []any) {
	infoLazyInt(instance.tag, dataProvider)
}

func (instance *logger) LogError(msg ...any) {
	errorInt(instance.tag, msg...)
}

func (instance *logger) LogIfError(when string, err error) {
	if err != nil {
		errorInt(when+":", errorInt)
	}
}

func (instance *logger) LogFatal(msg ...any) {
	fatalInt(instance.tag, msg...)
}
