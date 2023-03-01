package util

import (
	"reflect"
	"regexp"
	"strings"
)

type Generics string

func (gs Generics) GetList() []string {
	return strings.Split(string(gs), ",")
}

type TypeInfo struct {
	Path     string
	TypeName string
	Generics Generics
}

var matchGenerics *regexp.Regexp = regexp.MustCompile(`\[(.*)\]$`)
var matchTypeName *regexp.Regexp = regexp.MustCompile(`(.*?)\[(.*)`)

func GetTypeInfo(v any) *TypeInfo {
	rv := IdempotentReflectValueOf(v)
	var hasNameType reflect.Type
	if len(rv.Type().Name()) != 0 {
		hasNameType = rv.Type()
	} else {
		//todo slice type
		hasNameType = rv.Elem().Type()
	}
	fullName := hasNameType.Name()
	var (
		generics Generics
		name     string
	)
	names := matchGenerics.FindStringSubmatch(fullName)
	if names != nil {
		generics = Generics(names[1])
	}
	names = matchTypeName.FindStringSubmatch(fullName)
	if names != nil {
		name = names[1]
	} else {
		name = fullName
	}
	return &TypeInfo{
		Path:     hasNameType.PkgPath(),
		TypeName: name,
		Generics: generics,
	}
}
func IdempotentReflectValueOf(v any) reflect.Value {
	if val, itis := v.(reflect.Value); itis {
		return val
	}
	return reflect.ValueOf(v)
}
func GetUnderlying[T any](v any) T {
	return IdempotentReflectValueOf(v).Interface().(T)
}
func IsType[T any](v any, restrictGenerics ...bool) bool {
	var zero T
	givenTypeInfo := GetTypeInfo(v)
	expectedTypeInfo := GetTypeInfo(zero)
	shouldCompareGenerics := false
	if len(restrictGenerics) > 0 && restrictGenerics[0] {
		shouldCompareGenerics = true
	}
	return givenTypeInfo.Path == expectedTypeInfo.Path &&
		givenTypeInfo.TypeName == expectedTypeInfo.TypeName &&
		(!shouldCompareGenerics ||
			givenTypeInfo.Generics == expectedTypeInfo.Generics)
}
func IsInterface[T any](v any) bool {
	return IdempotentReflectValueOf(v).Type().Implements(reflect.TypeOf((*T)(nil)).Elem())
}
