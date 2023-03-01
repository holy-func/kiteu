package util

import "reflect"
type Callback func()
func NilDefend(vals ...any) {
	for _, v := range vals {
		if reflect.ValueOf(v).IsNil() {
			panic("unexpected nil val")
		}
	}
}
func IsSafeFunc(f any)bool{
	v:=reflect.ValueOf(f)
	return v.Kind()==reflect.Func&&!v.IsNil()
}
func SafeCallback(f Callback){
	if IsSafeFunc(f){
		reflect.ValueOf(f).Call(nil)
	}
}