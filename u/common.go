package u

import (
	"io"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](value T, err error) T {
	Must(err)
	return value
}

func Must3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2) {
	Must(err)
	return val1, val2
}

func Use[T io.Closer](resource T, block func(it T)) {
	defer CloseOptimistic(resource)
	block(resource)
}

func JoinAsString[T any](arr []T, converter func(val T) string, sep string) string {
	strArr := make([]string, len(arr))
	for i, val := range arr {
		strArr[i] = converter(val)
	}
	return strings.Join(strArr, sep)
}

func CloseOptimistic(resource io.Closer) {
	Must(resource.Close())
}

func WrapPanic[T any](block func() T, wrapperFunc func(reason any) any) T {
	defer func() {
		panicReason := recover()
		if panicReason != nil {
			panic(wrapperFunc(panicReason))
		}
	}()
	return block()
}
