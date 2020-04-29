package utils

import "reflect"

func OriginalType(typ reflect.Type) reflect.Type {
	switch typ.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		return OriginalType(typ.Elem())
	}
	return typ
}

func GetDereferenceType(typ reflect.Type) reflect.Type {
	switch typ.Kind() {
	case reflect.Ptr:
		return GetDereferenceType(typ.Elem())
	case reflect.Slice:
		return GetDereferenceType(typ.Elem())
	case reflect.Array:
		return GetDereferenceType(typ.Elem())
	case reflect.Chan:
		return GetDereferenceType(typ.Elem())
	case reflect.Map:
		return GetDereferenceType(typ.Elem())
	}
	return typ
}
