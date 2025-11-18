package ctx

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
)

type reflectiveServiceWrapper struct {
	sRef   any
	sValue reflect.Value
	sType  reflect.Type
	name   string
	ctx    context.Context
}

func newReflectiveServiceWrapper(ctx context.Context, service any, name string) *reflectiveServiceWrapper {
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
	return &reflectiveServiceWrapper{ctx: ctx, sRef: service, sValue: reflect.ValueOf(service).Elem(), sType: sTypeElem, name: sName}
}

func (w *reflectiveServiceWrapper) Name() string {
	return w.name
}

func (w *reflectiveServiceWrapper) init(serviceProvider ServiceProvider) error {
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
	injectContextFn := func(field reflect.StructField, value reflect.Value) {
		if field.Type.AssignableTo(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			setFieldValue(field, value, w.ctx)
		} else {
			logger.Fatal(w.name, "can't inject context.Context into field", field.Name)
		}
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

		if tag.context {
			injectContextFn(sField, sValue)
			continue
		}

		if tag.auto {
			// auto mode
			if sField.Type.AssignableTo(reflect.TypeOf((*logger.Logger)(nil)).Elem()) || sField.Type.AssignableTo(reflect.TypeOf((*slog.Logger)(nil))) {
				injectLoggerFn(sField, sValue, tag.name, tag.logAttrs)
			} else if sField.Type.AssignableTo(reflect.TypeOf((*context.Context)(nil)).Elem()) {
				injectContextFn(sField, sValue)
			} else {
				injectServiceFn(sField, sValue, tag.name)
			}
			continue
		}
	}

	InjectEnv(w.sRef)

	if v, ok := w.sRef.(Initializable); ok {
		if err := nopanic.Run(func() {
			v.Init(serviceProvider)
		}); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		}
	}
	if v, ok := w.sRef.(InitializableE); ok {
		if err := nopanic.RunE(func() error {
			return v.Init(serviceProvider)
		}); nopanic.IsPanicWrapperError(err) {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.(nopanic.PanicWrapperError).ToReasonError()
		} else if err != nil {
			return err
		}
	}
	if v, ok := w.sRef.(InitializableContext); ok {
		if err := nopanic.Run(func() {
			v.Init(w.ctx, serviceProvider)
		}); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		}
	}
	if v, ok := w.sRef.(InitializableContextE); ok {
		if err := nopanic.RunE(func() error {
			return v.Init(w.ctx, serviceProvider)
		}); nopanic.IsPanicWrapperError(err) {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.(nopanic.PanicWrapperError).ToReasonError()
		} else if err != nil {
			return err
		}
	}
	if v, ok := w.sRef.(Constructable); ok {
		if err := nopanic.Run(func() {
			v.Init()
		}); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		}
	}
	if v, ok := w.sRef.(ConstructableE); ok {
		if err := nopanic.RunE(func() error {
			return v.Init()
		}); nopanic.IsPanicWrapperError(err) {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.(nopanic.PanicWrapperError).ToReasonError()
		} else if err != nil {
			return err
		}
	}
	if v, ok := w.sRef.(ConstructableContext); ok {
		if err := nopanic.Run(func() {
			v.Init(w.ctx)
		}); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		}
	}
	if v, ok := w.sRef.(ConstructableContextE); ok {
		if err := nopanic.RunE(func() error {
			return v.Init(w.ctx)
		}); nopanic.IsPanicWrapperError(err) {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.(nopanic.PanicWrapperError).ToReasonError()
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (w *reflectiveServiceWrapper) afterStart() error {
	if v, ok := w.sRef.(StartAware); ok {
		logger.Debug(w.name, "is livecycle-aware, notify it for start event")
		if err := nopanic.Run(v.AfterStart); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (w *reflectiveServiceWrapper) beforeStop() error {
	if v, ok := w.sRef.(StopAware); ok {
		logger.Debug(w.name, "is livecycle-aware, notify it for stop event")
		if err := nopanic.Run(v.BeforeStop); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (w *reflectiveServiceWrapper) dispose() error {
	if v, ok := w.sRef.(Disposable); ok {
		if err := nopanic.Run(v.Dispose); err != nil {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.ToReasonError()
		}
	}
	if v, ok := w.sRef.(DisposableE); ok {
		if err := nopanic.RunE(v.Dispose); nopanic.IsPanicWrapperError(err) {
			logger.Debug(ctxTag, "panic:", err.Error())
			return err.(nopanic.PanicWrapperError).ToReasonError()
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (w *reflectiveServiceWrapper) unwrap() any {
	return w.sRef
}
