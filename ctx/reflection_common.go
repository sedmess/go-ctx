package ctx

import (
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

const (
	ctxReflectionTag     = "ctx"
	ctxEnvTag            = "env"
	ctxLoggerTag         = "logger"         // Deprecated: use ctxReflectionTag
	ctxImplementTag      = "implement"      // Deprecated: use ctxReflectionTag
	ctxImplementationTag = "implementation" // Deprecated: use ctxReflectionTag
	ctxInjectTag         = "inject"         // Deprecated: use ctxReflectionTag
	ctxEnvDefTag         = "envDef"         // Deprecated: use ctxEnvTag
)

var ctxReflectionProp = regexp.MustCompile("([a-z]+)(\\((.+)\\))?")

type reflectionTag struct {
	auto        bool
	inject      bool
	injectName  string
	impl        bool
	log         bool
	logName     string
	env         string
	envDef      bool
	envDefValue string
	name        string
}

func defineReflectionTag(tag reflect.StructTag) (rTag reflectionTag) {
	if ctxTag, ok := tag.Lookup(ctxReflectionTag); ok {
		if ctxTag == "" {
			rTag.auto = true
		} else {
			for _, field := range strings.FieldsFunc(ctxTag, func(r rune) bool {
				return r == ',' || r == ' ' || r == ':' || r == ';'
			}) {
				submatch := ctxReflectionProp.FindStringSubmatch(field)
				if len(submatch) <= 1 {
					continue
				}
				switch submatch[1] {
				case "inject":
					rTag.inject = true
					if len(submatch) == 4 {
						rTag.injectName = submatch[3]
					}
				case "impl":
					rTag.impl = true
				case "logger":
					rTag.log = true
					if len(submatch) == 4 {
						rTag.logName = submatch[3]
					}
				default:
					rTag.auto = true
					rTag.name = field
				}
			}
		}
	} else {
		if loggerTag, ok := tag.Lookup(ctxLoggerTag); ok {
			rTag.log = true
			rTag.logName = loggerTag
		}

		if hasTag(tag, ctxImplementTag) || hasTag(tag, ctxImplementationTag) {
			rTag.impl = true
		}
		if injectTag, ok := tag.Lookup(ctxInjectTag); ok {
			rTag.inject = true
			rTag.injectName = injectTag
		}
	}
	if envTag, ok := tag.Lookup(ctxEnvTag); ok {
		envName, envValue, found := strings.Cut(envTag, "=")
		rTag.env = envName
		if found {
			rTag.envDef = true
			rTag.envDefValue = envValue
		} else {
			if defTag, ok := tag.Lookup(ctxEnvDefTag); ok {
				rTag.envDef = true
				rTag.envDefValue = defTag
			}
		}
	}
	return rTag
}

func hasTag(sTag reflect.StructTag, tagName string) bool {
	_, ok := sTag.Lookup(tagName)
	return ok
}

func setFieldValue(f reflect.StructField, v reflect.Value, value any) {
	if !f.IsExported() {
		v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()

	}
	v.Set(reflect.ValueOf(value))
}

func DefineServiceName(service any) string {
	var sName string
	sType := reflect.TypeOf(service)
	sTypeElem := sType.Elem()

	if v, ok := service.(Named); ok {
		sName = v.Name()
	} else {
		var nameCandidateField *reflect.StructField
		for i := 0; i < sTypeElem.NumField(); i++ {
			sField := sTypeElem.Field(i)
			if sField.Anonymous && sField.Type.Kind() == reflect.Interface {
				tag := defineReflectionTag(sField.Tag)
				if tag.impl {
					if nameCandidateField == nil {
						nameCandidateField = &sField
					} else {
						nameCandidateField = nil
						break
					}
				}
			}
		}
		if nameCandidateField != nil {
			sName = nameCandidateField.Type.String()
		} else {
			sName = sType.String()
		}
	}
	return sName
}
