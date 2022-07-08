package ctx

import (
	"bufio"
	"os"
	"strconv"
	"strings"
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
		substrs := strings.Split(scanner.Text(), "=")
		if len(substrs) < 2 {
			continue
		}
		envFileMap[substrs[0]] = substrs[1]
	}
	properties = envFileMap
}

type EnvValue struct {
	name  string
	value string
}

func (instance EnvValue) IsPresent() bool {
	return len(instance.value) != 0
}

func (instance EnvValue) AsString() string {
	instance.fatalIfNotExists()
	return instance.value
}

func (instance EnvValue) AsStringDefault(def string) string {
	if instance.IsPresent() {
		return instance.value
	} else {
		return def
	}
}

func (instance EnvValue) AsStringArray() []string {
	return strings.Split(instance.value, ",")
}

func (instance EnvValue) AsInt() int {
	instance.fatalIfNotExists()
	if a, err := strconv.Atoi(instance.value); err != nil {
		panic(instance.name + ": can't convert to integer: " + instance.value)
		return 0
	} else {
		return a
	}
}

func (instance EnvValue) AsIntDefault(def int) int {
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

func (instance EnvValue) AsIntArray() []int {
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

func (instance EnvValue) AsIntArrayDefault() []int {
	if instance.IsPresent() {
		return instance.AsIntArray()
	} else {
		return make([]int, 0)
	}
}

func (instance EnvValue) AsBool() bool {
	instance.fatalIfNotExists()
	if boolValue, err := strconv.ParseBool(instance.value); err != nil {
		panic(instance.name + ": can't convert to boolean: " + instance.value)
		return false
	} else {
		return boolValue
	}
}

func (instance EnvValue) AsBoolDefault(def bool) bool {
	if instance.IsPresent() {
		return instance.AsBool()
	} else {
		return def
	}
}

func (instance EnvValue) fatalIfNotExists() {
	if !instance.IsPresent() {
		panic("environment variable " + instance.name + " not set")
	}
}

func GetEnv(name string) EnvValue {
	var value string
	value = os.Getenv(name)
	if len(value) == 0 && properties != nil {
		value = properties[name]
	}
	return EnvValue{name: name, value: value}
}
