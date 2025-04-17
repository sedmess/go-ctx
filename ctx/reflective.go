package ctx

import (
	"github.com/sedmess/go-ctx/ctx/logger"
	"log/slog"
	"reflect"
)

type serviceWrapper interface {
	service() any
}

type reflectiveServiceWrapper struct {
	sRef   any
	sValue reflect.Value
	sType  reflect.Type
	name   string
}

func unwrap(service Service) any {
	if s, ok := service.(serviceWrapper); ok {
		return s.service()
	} else {
		panic("unexpected error")
	}
}

func newReflectiveServiceWrapper(service any, name string) *reflectiveServiceWrapper {
	sType := reflect.TypeOf(service)
	if sType.Kind() != reflect.Pointer {
		logger.Fatal(ctxTag, "["+sType.String()+"] can't be a service - must be a pointer to a struct")
	}
	sTypeElem := sType.Elem()
	if sTypeElem.Kind() != reflect.Struct {
		logger.Fatal(ctxTag, "["+sType.String()+"] can't be a service - must be a pointer to a struct")
	}
	var sName string
	if name == "" {
		sName = DefineServiceName(service)
	} else {
		sName = name
	}
	return &reflectiveServiceWrapper{sRef: service, sValue: reflect.ValueOf(service).Elem(), sType: sTypeElem, name: sName}
}

func (w *reflectiveServiceWrapper) Init(serviceProvider ServiceProvider) {
	injectLoggerFn := func(field reflect.StructField, value reflect.Value, name string, attrs [][]string) {
		var slogAttrs []any
		for _, attr := range attrs {
			slogAttrs = append(slogAttrs, slog.String(attr[0], attr[1]))
		}
		if field.Type.AssignableTo(reflect.TypeOf((*logger.Logger)(nil)).Elem()) {
			var l logger.Logger
			if name != "" {
				logger.Debug(w.name, "inject logger \""+name+"\" into field", field.Name)
				l = logger.New(name, slogAttrs...)
			} else {
				logger.Debug(w.name, "inject default logger into field", field.Name)
				l = logger.New(w.name, slogAttrs...)
			}
			setFieldValue(field, value, l)
		} else if field.Type.AssignableTo(reflect.TypeOf((*slog.Logger)(nil))) {
			var l *slog.Logger
			if name != "" {
				logger.Debug(w.name, "inject slogger \""+name+"\" into field", field.Name)
				l = logger.CreateSlogFor(name, slogAttrs...)
			} else {
				logger.Debug(w.name, "inject default slogger into field", field.Name)
				l = logger.CreateSlogFor(w.name, slogAttrs...)
			}
			setFieldValue(field, value, l)
		} else {
			logger.Fatal(w.name, "can't inject logger into field", field.Name)
		}
	}
	injectServiceFn := func(field reflect.StructField, value reflect.Value, name string) {
		var service any
		if name != "" {
			logger.Debug(w.name, "lookup dependency", name, "for field", field.Name)
			service = serviceProvider.ByName(name)
		} else {
			logger.Debug(w.name, "lookup dependency of type", field.Type.String(), "for field", field.Name)
			service = serviceProvider.byReflectType(field.Type)
		}
		setFieldValue(field, value, service)
	}

	for i := 0; i < w.sType.NumField(); i++ {
		sField := w.sType.Field(i)
		sValue := w.sValue.Field(i)

		tag := defineReflectionTag(sField.Tag)

		if tag.log {
			injectLoggerFn(sField, sValue, tag.logName, tag.logAttrs)
			continue
		}

		if tag.inject {
			injectServiceFn(sField, sValue, tag.injectName)
			continue
		}

		if tag.auto {
			// auto mode
			if sField.Type.AssignableTo(reflect.TypeOf((*logger.Logger)(nil)).Elem()) || sField.Type.AssignableTo(reflect.TypeOf((*slog.Logger)(nil))) {
				injectLoggerFn(sField, sValue, tag.name, tag.logAttrs)
			} else {
				injectServiceFn(sField, sValue, tag.name)
			}
			continue
		}
	}

	InjectEnv(w.sRef)

	if v, ok := w.sRef.(Initializable); ok {
		v.Init(serviceProvider)
	}
	if v, ok := w.sRef.(Constructable); ok {
		v.Init()
	}
}

func (w *reflectiveServiceWrapper) Name() string {
	return w.name
}

func (w *reflectiveServiceWrapper) AfterStart() {
	if v, ok := w.sRef.(StartAware); ok {
		v.AfterStart()
	}
}

func (w *reflectiveServiceWrapper) BeforeStop() {
	if v, ok := w.sRef.(StopAware); ok {
		v.BeforeStop()
	}
}

func (w *reflectiveServiceWrapper) Dispose() {
	if v, ok := w.sRef.(Disposable); ok {
		v.Dispose()
	}
}

func (w *reflectiveServiceWrapper) service() any {
	return w.sRef
}
