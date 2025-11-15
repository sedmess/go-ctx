package ctx

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sedmess/go-ctx/ctx/logger"
)

const defaultPropertiesFileName = ".env"
const defaultCustomPropertiesFileName = ".env_custom"

var propertiesOnce sync.Once
var properties map[string]string

func initProperties() {
	propertiesOnce.Do(func() {
		envFileMap := make(map[string]string)

		appDefaultProperties.Range(func(key, value any) bool {
			envFileMap[key.(string)] = value.(string)
			return true
		})

		readFile(defaultPropertiesFileName, envFileMap)
		readFile(defaultCustomPropertiesFileName, envFileMap)
		readArgs(envFileMap)

		properties = envFileMap
	})
}

func readArgs(properties map[string]string) {
	for _, arg := range os.Args[1:] {
		if a, found := strings.CutPrefix(arg, "--"); found {
			substrs := strings.SplitN(a, "=", 2)
			if len(substrs) < 2 {
				continue
			}
			properties[strings.ToUpper(substrs[0])] = substrs[1]
		}
	}
}

func readFile(path string, properties map[string]string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		substrs := strings.SplitN(scanner.Text(), "=", 2)
		if len(substrs) < 2 {
			continue
		}
		properties[strings.ToUpper(substrs[0])] = substrs[1]
	}
}

var appDefaultProperties sync.Map

func SetEnv(key string, value string) {
	appDefaultProperties.Store(key, value)
}

var envTypes = map[reflect.Type]func(e *EnvValue) any{
	reflect.TypeOf(""): func(e *EnvValue) any {
		return e.AsString()
	},
	reflect.TypeOf(0): func(e *EnvValue) any {
		return e.AsInt()
	},
	reflect.TypeOf(true): func(e *EnvValue) any {
		return e.AsBool()
	},
	reflect.TypeOf(int64(0)): func(e *EnvValue) any {
		return e.AsInt64()
	},
	reflect.TypeOf(make([]string, 0)): func(e *EnvValue) any {
		return e.AsStringArrayDefault([]string{})
	},
	reflect.TypeOf(make(map[string]bool)): func(e *EnvValue) any {
		return e.AsStringSetDefault([]string{})
	},
	reflect.TypeOf(make([]int, 0)): func(e *EnvValue) any {
		return e.AsIntArrayDefault()
	},
	reflect.TypeOf(make(map[int]bool)): func(e *EnvValue) any {
		return e.AsIntSetDefault()
	},
	reflect.TypeOf(make([]int64, 0)): func(e *EnvValue) any {
		return e.AsInt64ArrayDefault()
	},
	reflect.TypeOf(make(map[int64]bool)): func(e *EnvValue) any {
		return e.AsInt64SetDefault()
	},
	reflect.TypeOf(time.Second): func(e *EnvValue) any {
		return e.AsDuration()
	},
	reflect.TypeOf(time.Unix(0, 0)): func(e *EnvValue) any { return e.AsTime() },
	reflect.TypeOf(make(map[string]*EnvValue)): func(e *EnvValue) any {
		return e.AsMapDefault()
	},
}

type EnvValue struct {
	name  string
	set   bool
	value string
}

func (instance *EnvValue) Name() string {
	return instance.name
}

func (instance *EnvValue) IsPresent() bool {
	return instance.set
}

func (instance *EnvValue) AsString() string {
	instance.fatalIfNotExists()
	return instance.value
}

func (instance *EnvValue) AsStringDefault(def string) string {
	if instance.IsPresent() {
		return instance.value
	} else {
		return def
	}
}

func (instance *EnvValue) AsStringArray() []string {
	instance.fatalIfNotExists()
	return strings.Split(instance.value, ",")
}

func (instance *EnvValue) AsStringArrayDefault(def []string) []string {
	if instance.IsPresent() {
		return strings.Split(instance.value, ",")
	} else {
		return def
	}
}

func (instance *EnvValue) AsStringSet() map[string]bool {
	arr := instance.AsStringArray()
	set := make(map[string]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsStringSetDefault(def []string) map[string]bool {
	arr := instance.AsStringArrayDefault(def)
	set := make(map[string]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsInt() int {
	instance.fatalIfNotExists()
	if a, err := strconv.Atoi(instance.value); err != nil {
		panic(instance.name + ": can't convert to integer: " + instance.value)
		return 0
	} else {
		return a
	}
}

func (instance *EnvValue) AsIntDefault(def int) int {
	if instance.IsPresent() {
		if a, err := strconv.Atoi(instance.value); err != nil {
			panic(instance.name + ": can't convert to integer: " + instance.value)
			return 0
		} else {
			return a
		}
		return 0
	} else {
		return def
	}
}

func (instance *EnvValue) AsIntArray() []int {
	instance.fatalIfNotExists()
	strs := instance.AsStringArray()
	ints := make([]int, len(strs))
	for i := range strs {
		if a, err := strconv.Atoi(strs[i]); err != nil {
			panic(instance.name + ": can't convert to integer: " + strs[i])
		} else {
			ints[i] = a
		}
	}
	return ints
}

func (instance *EnvValue) AsIntArrayDefault() []int {
	if instance.IsPresent() {
		return instance.AsIntArray()
	} else {
		return make([]int, 0)
	}
}

func (instance *EnvValue) AsIntSet() map[int]bool {
	arr := instance.AsIntArray()
	set := make(map[int]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsIntSetDefault() map[int]bool {
	if !instance.IsPresent() {
		return make(map[int]bool)
	}
	arr := instance.AsIntArray()
	set := make(map[int]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsInt64() int64 {
	instance.fatalIfNotExists()
	if a, err := strconv.ParseInt(instance.value, 10, 64); err != nil {
		panic(instance.name + ": can't convert to int64: " + instance.value)
		return 0
	} else {
		return a
	}
}

func (instance *EnvValue) AsInt64Default(def int64) int64 {
	if instance.IsPresent() {
		if a, err := strconv.ParseInt(instance.value, 10, 64); err != nil {
			panic(instance.name + ": can't convert to int64: " + instance.value)
			return 0
		} else {
			return a
		}
		return 0
	} else {
		return def
	}
}

func (instance *EnvValue) AsInt64Array() []int64 {
	instance.fatalIfNotExists()
	strs := instance.AsStringArray()
	ints := make([]int64, len(strs))
	for i := range strs {
		if a, err := strconv.ParseInt(strs[i], 10, 64); err != nil {
			panic(instance.name + ": can't convert to int64: " + strs[i])
		} else {
			ints[i] = a
		}
	}
	return ints
}

func (instance *EnvValue) AsInt64ArrayDefault() []int64 {
	if instance.IsPresent() {
		return instance.AsInt64Array()
	} else {
		return make([]int64, 0)
	}
}

func (instance *EnvValue) AsInt64Set() map[int64]bool {
	arr := instance.AsInt64Array()
	set := make(map[int64]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsInt64SetDefault() map[int64]bool {
	if !instance.IsPresent() {
		return make(map[int64]bool)
	}
	arr := instance.AsInt64Array()
	set := make(map[int64]bool)
	for _, v := range arr {
		set[v] = true
	}
	return set
}

func (instance *EnvValue) AsBool() bool {
	instance.fatalIfNotExists()
	if boolValue, err := strconv.ParseBool(instance.value); err != nil {
		panic(instance.name + ": can't convert to boolean: " + instance.value)
		return false
	} else {
		return boolValue
	}
}

func (instance *EnvValue) AsBoolDefault(def bool) bool {
	if instance.IsPresent() {
		return instance.AsBool()
	} else {
		return def
	}
}

func (instance *EnvValue) AsDuration() time.Duration {
	instance.fatalIfNotExists()
	if durationValue, err := time.ParseDuration(instance.value); err != nil {
		panic(instance.name + ": can't convert to time.Duration: " + instance.value)
		return 0
	} else {
		return durationValue
	}
}

func (instance *EnvValue) AsDurationDefault(def time.Duration) time.Duration {
	if instance.IsPresent() {
		return instance.AsDuration()
	} else {
		return def
	}
}

func (instance *EnvValue) AsTime() time.Time {
	instance.fatalIfNotExists()
	if val, err := time.Parse(time.RFC3339, instance.value); err != nil {
		panic(instance.name + ": can't convert to time.Time using RFC3339 format: " + instance.value)
		return time.Now()
	} else {
		return val
	}
}

func (instance *EnvValue) AsTimeDefault(def time.Time) time.Time {
	if instance.IsPresent() {
		return instance.AsTime()
	} else {
		return def
	}
}

func (instance *EnvValue) AsMap() map[string]*EnvValue {
	instance.fatalIfNotExists()
	result := make(map[string]*EnvValue)
	for _, str := range strings.Split(instance.value, "|") {
		parts := strings.Split(str, "=")
		if len(parts) != 2 {
			panic(instance.name + ": can't find key-value pair in part \"" + str + "\"")
		}
		result[parts[0]] = &EnvValue{
			name:  instance.name + "(map)." + parts[0],
			set:   true,
			value: parts[1],
		}
	}
	return result
}

func (instance *EnvValue) AsMapDefault() map[string]*EnvValue {
	if !instance.IsPresent() {
		return make(map[string]*EnvValue)
	}
	result := make(map[string]*EnvValue)
	for _, str := range strings.Split(instance.value, "|") {
		parts := strings.Split(str, "=")
		if len(parts) != 2 {
			panic(instance.name + ": can't find key-value pair in part \"" + str + "\"")
		}
		result[parts[0]] = &EnvValue{
			name:  instance.name + "(map)." + parts[0],
			set:   true,
			value: parts[1],
		}
	}
	return result
}

func (instance *EnvValue) asType(rType reflect.Type, hasDef bool, def string) (any, bool) {
	if hasDef && !instance.IsPresent() {
		instance.set = true
		instance.value = def
	}
	if fn, found := envTypes[rType]; found {
		return fn(instance), true
	} else {
		return nil, false
	}
}

func (instance *EnvValue) fatalIfNotExists() {
	if !instance.IsPresent() {
		panic("environment variable " + instance.name + " not set")
	}
}

func (instance *EnvValue) String() string {
	if instance.IsPresent() {
		return "EnvValue: (" + instance.name + ": " + instance.value + ")"
	} else {
		return "EnvValue: (" + instance.name + ": {not set}" + ")"
	}
}

func GetEnv(name string) *EnvValue {
	initProperties()
	var value string
	var set bool
	value, set = os.LookupEnv(name)
	if !set && properties != nil {
		value, set = properties[name]
	}
	return &EnvValue{name: name, set: set, value: value}
}

func GetEnvCustom(custom string, name string) *EnvValue {
	return getEnvCustom(custom, name, false)
}

//goland:noinspection ALL
func GetEnvCustomOrDefault(custom string, name string) *EnvValue {
	return getEnvCustom(custom, name, true)
}

func getEnvCustom(custom string, name string, allowDefault bool) *EnvValue {
	key := custom + "_" + name
	env := GetEnv(key)
	if env.IsPresent() {
		logger.Debug(ctxTag, "get", name, "customized by", custom, "as", key)
		return env
	} else if allowDefault {
		logger.Debug(ctxTag, "get", name, "customized by", custom, "as", name)
		return GetEnv(name)
	} else {
		return env
	}
}

func Env[T any]() T {
	t := new(T)
	InjectEnv(t)
	return *t
}

func InjectEnv(target any) {
	sType := reflect.TypeOf(target)
	if sType.Kind() != reflect.Pointer {
		logger.Fatal(ctxTag, sType.String(), "must be pointer to a struct")
	}
	sTypeElem := sType.Elem()
	if sTypeElem.Kind() != reflect.Struct {
		logger.Fatal(ctxTag, sType.String(), "must be pointer to a struct")
	}
	sValueElem := reflect.ValueOf(target).Elem()

	for i := 0; i < sTypeElem.NumField(); i++ {
		sField := sTypeElem.Field(i)
		sValue := sValueElem.Field(i)
		sFieldType := sField.Type

		tag := defineReflectionTag(sField.Tag)
		if tag.env != "" {
			env := GetEnv(tag.env)
			if sField.Type.AssignableTo(reflect.TypeOf((*EnvValue)(nil))) {
				logger.Debug(ctxTag, "inject EnvValue", tag.env, "into", sType.String()+"."+sField.Name)
				setFieldValue(sField, sValue, env)
			} else {
				if eValue, ok := env.asType(sFieldType, tag.envDef, tag.envDefValue); ok {
					logger.Debug(ctxTag, "inject EnvValue", tag.env, "into", sType.String()+"."+sField.Name, "with type", sFieldType.String())
					setFieldValue(sField, sValue, eValue)
				} else {
					logger.Fatal(ctxTag, "can't inject EnvValue", tag.env, "into", sType.String()+"."+sField.Name, "with type", sFieldType.String(), "- type not supported")
				}
			}
			continue
		}
	}
}
