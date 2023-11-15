package ctx

import (
	"github.com/sedmess/go-ctx/logger"
	"reflect"
	"unsafe"
)

const tagLogger = "logger"
const tagEnv = "env"
const tagInject = "inject"

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
		return service
	}
}

func newReflectiveServiceWrapper(service any) *reflectiveServiceWrapper {
	var sName string
	if v, ok := service.(Named); ok {
		sName = v.Name()
	} else {
		sName = reflect.TypeOf(service).String()
	}
	return &reflectiveServiceWrapper{sRef: service, sValue: reflect.ValueOf(service).Elem(), sType: reflect.TypeOf(service).Elem(), name: sName}
}

func (w *reflectiveServiceWrapper) Init(serviceProvider ServiceProvider) {
	for i := 0; i < w.sType.NumField(); i++ {
		sField := w.sType.Field(i)
		sValue := w.sValue.Field(i)

		sFieldType := sField.Type

		value, ok := sField.Tag.Lookup(tagLogger)
		if ok {
			logger.Debug(w.name, "inject logger into field", sField.Name)
			var l logger.Logger
			if value != "" {
				l = logger.NewWithTag(value)
			} else {
				l = logger.New(w)
			}
			setFieldValue(sField, sValue, l)
			continue
		}

		value, ok = sField.Tag.Lookup(tagEnv)
		if ok && value != "" {
			env := GetEnv(value)
			if sField.Type.AssignableTo(reflect.TypeOf((*EnvValue)(nil))) {
				logger.Debug(w.name, "inject EnvValue", value, "into field", sField.Name)
				setFieldValue(sField, sValue, env)
			} else {
				if eValue, ok := env.asType(sFieldType); ok {
					logger.Debug(w.name, "inject EnvValue", value, "into field", sField.Name, "with type", sFieldType.String())
					setFieldValue(sField, sValue, eValue)
				} else {
					logger.Fatal(w.name, "can't inject EnvValue", value, "into field", sField.Name, "with type", sFieldType.String(), "- type not supported")
				}
			}
			continue
		}

		value, ok = sField.Tag.Lookup(tagInject)
		if ok {
			var service any
			if value != "" {
				logger.Debug(w.name, "lookup dependency", value, "for field", sField.Name)
				service = serviceProvider.ByName(value)
			} else {
				logger.Debug(w.name, "lookup dependency of type", sFieldType.String(), "for field", sField.Name)
				service = serviceProvider.byReflectType(sFieldType)
			}
			setFieldValue(sField, sValue, service)
			continue
		}
	}

	if v, ok := w.sRef.(Initializable); ok {
		v.Init(serviceProvider)
	}
}

func (w *reflectiveServiceWrapper) Name() string {
	return w.name
}

func (w *reflectiveServiceWrapper) AfterStart() {
	if v, ok := w.sRef.(LifecycleAware); ok {
		v.AfterStart()
	}
}

func (w *reflectiveServiceWrapper) BeforeStop() {
	if v, ok := w.sRef.(LifecycleAware); ok {
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

func setFieldValue(f reflect.StructField, v reflect.Value, value any) {
	if !f.IsExported() {
		v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()

	}
	v.Set(reflect.ValueOf(value))
}
