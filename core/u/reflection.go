package u

import (
	"reflect"
)

func GetInterfaceName[T any]() string {
	return reflect.TypeOf((*T)(nil)).Elem().String()
}
