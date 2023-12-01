package ctx

import (
	"bufio"
	"github.com/sedmess/go-ctx/logger"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const defaultPropertiesFileName = ".env"

var properties map[string]string

func init() {
	file, err := os.Open(defaultPropertiesFileName)
	if err != nil {
		properties = nil
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	envFileMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		substrs := strings.SplitN(scanner.Text(), "=", 2)
		if len(substrs) < 2 {
			continue
		}
		envFileMap[substrs[0]] = substrs[1]
	}
	properties = envFileMap
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
	reflect.TypeOf(make(map[string]*EnvValue)): func(e *EnvValue) any {
		return e.AsMapDefault()
	},
}

type EnvValue struct {
	name  string
	value string
}

func (instance *EnvValue) Name() string {
	return instance.name
}

func (instance *EnvValue) IsPresent() bool {
	return len(instance.value) != 0
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
			value: parts[1],
		}
	}
	return result
}

func (instance *EnvValue) asType(rType reflect.Type, def string) (any, bool) {
	if def != "" && !instance.IsPresent() {
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
	var value string
	value = os.Getenv(name)
	if len(value) == 0 && properties != nil {
		value = properties[name]
	}
	return &EnvValue{name: name, value: value}
}

func GetEnvCustom(custom string, name string) *EnvValue {
	return getEnvCustom(custom, name, false)
}

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
