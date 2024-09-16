package ctx

import (
	"reflect"
	"unsafe"
)

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
				if _, ok := sField.Tag.Lookup(tagImplement); ok {
					if nameCandidateField == nil {
						nameCandidateField = &sField
					} else {
						nameCandidateField = nil
						break
					}
				}
				if _, ok := sField.Tag.Lookup(tagImplementation); ok {
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
