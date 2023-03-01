package arrayModel

import (
	"fmt"
	"sort"
)

type Compare[T any] struct {
	Less  func(T, T) bool
	Equal func(T, T) bool
}
type JSArray[T any] struct {
	value   []T
	compare *Compare[T]
}
type TestFunc[T any] func(value T, index int) bool

type Callback[T any] func(T, int)

type ArrayModel[T any] interface {
	Get() []T
	Set([]T)
	Push(...T) int
	Shift() T
	Pop() T
	Unshift(...T) int
	Every(TestFunc[T]) bool
	Some(TestFunc[T]) bool
	Each(Callback[T])
	Includes(T) bool
	FindIndex(TestFunc[T]) int
	Length() int
	Filter(TestFunc[T]) ArrayModel[T]
}

func (arr *JSArray[T]) Get() []T {
	return genVice(arr.value)
}
func (arr *JSArray[T]) Set(val []T) {
	arr.value = val
}
func genVice[T any](src []T) []T {
	vice := make([]T, len(src))
	copy(vice, src)
	return vice
}

func Array[T any](arr []T, compare *Compare[T]) ArrayModel[T] {
	if arr == nil {
		arr = []T{}
	}
	array := &JSArray[T]{arr, compare}
	return array
}
func (arr *JSArray[T]) Push(items ...T) int {
	arr.Set(append(arr.value, items...))
	return len(arr.value)
}
func (arr *JSArray[T]) String() string {
	return fmt.Sprint(arr.Get())
}

func (arr *JSArray[T]) Reverse() *JSArray[T] {
	n := len(arr.value)
	ret := make([]T, n)
	for i, v := range arr.value {
		ret[n-i-1] = v
	}
	arr.Set(ret)
	return arr
}

func (arr *JSArray[T]) Pop() T {
	n := len(arr.value)
	pop := arr.value[n-1]
	arr.Set(arr.value[:n-1])
	return pop
}

func (arr *JSArray[T]) Length() int {
	return len(arr.Get())
}

func (arr *JSArray[T]) Sort(lessers ...func(prev, next T) bool) ArrayModel[T] {
	less := arr.compare.Less
	if len(lessers) == 1 {
		less = lessers[0]
	}
	if less == nil {
		panic("need less func")
	}
	newArr := genVice(arr.value)
	sort.Slice(newArr, func(i, j int) bool {
		return less(newArr[i], newArr[j])
	})
	arr.Set(newArr)
	return arr
}
func (arr *JSArray[T]) Concat(arrs ...*JSArray[T]) ArrayModel[T] {
	ret := Array(nil, arr.compare)
	value := []T{}
	for _, array := range arrs {
		value = append(value, array.Get()...)
	}
	ret.Set(value)
	return ret
}
func (arr *JSArray[T]) Shift() T {
	shift := arr.value[0]
	arr.Set(arr.value[1:])
	return shift
}
func (arr *JSArray[T]) Unshift(items ...T) int {
	arr.Set(append(items, arr.Get()...))
	return len(arr.value)
}
func (arr *JSArray[T]) Some(testFunc TestFunc[T]) bool {
	value := arr.Get()
	for i, v := range value {
		if testFunc(v, i) {
			return true
		}
	}
	return false
}
func (arr *JSArray[T]) Join(separator string) string {
	value := arr.Get()
	vars := make([]any, len(value))
	formatStr := ""
	for i, v := range value {
		vars[i] = v
		if i != 0 {
			formatStr += separator + "%v"
		} else {
			formatStr += "%v"
		}
	}
	return fmt.Sprintf(formatStr, vars...)
}

func (arr *JSArray[T]) Each(cb Callback[T]) {
	for i, v := range arr.Get() {
		cb(v, i)
	}
}

func (arr *JSArray[T]) At(pos int) T {
	value := arr.Get()
	if pos >= 0 {
		return value[pos]
	} else {
		return value[len(value)+pos]
	}
}
func (arr *JSArray[T]) Slice(params ...int) ArrayModel[T] {
	value := arr.Get()
	n := len(params)
	if n == 0 {
		return Array(genVice(value), arr.compare)
	} else if n == 1 {
		return Array(genVice(value[params[0]:]), arr.compare)
	} else {
		return Array(genVice(value[params[0]:params[1]]), arr.compare)
	}
}

func (arr *JSArray[T]) Map(mapFunc func(T, int) any) ArrayModel[any] {
	value := arr.Get()
	ret := make([]any, len(value))
	for i, v := range value {
		ret[i] = mapFunc(v, i)
	}
	return Array(ret, nil)
}

func (arr *JSArray[T]) Every(testFunc TestFunc[T]) bool {
	value := arr.Get()
	for i, v := range value {
		if !testFunc(v, i) {
			return false
		}
	}
	return true
}

func (arr *JSArray[T]) Includes(t T) bool {
	if arr.compare.Equal == nil {
		panic("need equal func")
	}
	value := arr.Get()
	for _, v := range value {
		if arr.compare.Equal(v, t) {
			return true
		}
	}
	return false
}

func (arr *JSArray[T]) FindIndex(testFunc TestFunc[T]) int {
	value := arr.Get()
	for i, v := range value {
		if testFunc(v, i) {
			return i
		}
	}
	return -1
}

func (arr *JSArray[T]) Filter(testFunc TestFunc[T]) ArrayModel[T] {
	ret := Array(nil, arr.compare)
	value := arr.Get()
	for i, v := range value {
		if testFunc(v, i) {
			ret.Push(v)
		}
	}
	return ret
}

func (arr *JSArray[T]) Reduce(reducer func(prev, next any) any, initial any) any {
	value := arr.Get()
	ret := initial
	for _, v := range value {
		ret = reducer(ret, v)
	}
	return ret
}

func Type[T any](arr *JSArray[any]) ArrayModel[T] {
	value := arr.Get()
	ret := make([]T, len(value))
	for i, v := range value {
		ret[i] = v.(T)
	}
	return Array(ret, &Compare[T]{
		Less:  func(t1, t2 T) bool { return arr.compare.Less(t1, t2) },
		Equal: func(t1, t2 T) bool { return arr.compare.Equal(t1, t2) },
	})
}
