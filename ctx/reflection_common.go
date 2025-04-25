package ctx

import (
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

const (
	ctxReflectionTag = "ctx"
	ctxEnvTag        = "env"
)

var ctxReflectionProp = regexp.MustCompile("([A-z]+)(\\((.+)\\))?")

type reflectionTag struct {
	auto        bool
	inject      bool
	injectName  string
	impl        bool
	log         bool
	logName     string
	logAttrs    [][]string
	env         string
	envDef      bool
	envDefValue string
	context     bool
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
				case "loggerAttr":
					if len(submatch) == 4 {
						attr := submatch[3]
						if key, value, found := strings.Cut(attr, "="); found {
							rTag.logAttrs = append(rTag.logAttrs, []string{key, value})
						}
					}
				case "context":
					rTag.context = true
				default:
					rTag.auto = true
					rTag.name = field
				}
			}
		}
	}
	if envTag, ok := tag.Lookup(ctxEnvTag); ok {
		envName, envValue, found := strings.Cut(envTag, "=")
		rTag.env = envName
		if found {
			rTag.envDef = true
			rTag.envDefValue = envValue
		}
	}
	return rTag
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
