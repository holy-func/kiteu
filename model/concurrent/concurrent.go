package concurrentModel

import (
	"fmt"
	lockModel "kiteu/model/lock"
	"kiteu/model/new"
	"kiteu/util"
	"sync"
)

var defaultCache int = 2

func SetDefaultCache(n int) {
	defaultCache = n
}

type ConcurrentStore[T any, Data any] interface {
	GetData() T
	GetCache() int
	AddData(Data)
}

type ConcurrentModel[T any, Data any] interface {
	GetData() T
	AddData(Data)
	Wait()
	Seal()
	IsSealed() bool
	String() string
}

type ConcurrentData[T any, Data any] struct {
	ch       chan Data
	lock     chan byte
	isSealed bool
	once     sync.Once
	store    ConcurrentStore[T, Data]
	newModel.Init
	lockModel.RWLock
}

func (cd *ConcurrentData[T, Data]) new(store ConcurrentStore[T, Data]) *ConcurrentData[T, Data] {
	cd.InitCall(func() {
		cd.store = store
		cache := store.GetCache()
		cd.ch = make(chan Data, cache)
		cd.lock = make(chan byte)
		go func() {
			for v := range cd.ch {
				store.AddData(v)
			}
			close(cd.lock)
		}()
	})
	return cd
}
func (cd *ConcurrentData[T, Data]) GetData() T {
	<-cd.lock
	return cd.store.GetData()
}
func (cd *ConcurrentData[T, Data]) String() string {
	return fmt.Sprintf(`{ data %v, sealed %v }`, cd.store.GetData(), cd.isSealed)
}
func (cd *ConcurrentData[T, Data]) AddData(data Data) {
	cd.trydo(func() {
		cd.ch <- data
	})
}
func (cd *ConcurrentData[T, Data]) Seal() {
	cd.once.Do(func() {
		cd.Write(func() {
			cd.isSealed = true
			close(cd.ch)
			<-cd.lock
		})
	})
}
func (cd *ConcurrentData[T, Data]) IsSealed() bool {
	return cd.trydo(nil)
}
func (cd *ConcurrentData[T, Data]) trydo(f util.Callback) bool {
	var sealed bool
	cd.Write(func() {
		sealed = cd.isSealed
		if !sealed {
			util.SafeCallback(f)
		}
	})
	return sealed
}
func (cd *ConcurrentData[T, Data]) Wait() {
	<-cd.lock
}

type ConcurrentSliceStore[Data any] struct {
	catch int
	store []Data
}

func (cs *ConcurrentSliceStore[T]) GetData() []T {
	return cs.store
}
func (cs *ConcurrentSliceStore[T]) GetCache() int {
	return cs.catch
}
func (cs *ConcurrentSliceStore[T]) AddData(data T) {
	cs.store = append(cs.store, data)
}

func ConcurrentSlice[T any](optionalCache ...int) ConcurrentStore[[]T, T] {
	catch := defaultCache
	if len(optionalCache) != 0 {
		catch = optionalCache[0]
	}
	return &ConcurrentSliceStore[T]{
		catch: catch,
		store: []T{},
	}
}

func CreateConcurrent[Data any, T any](cc ConcurrentStore[T, Data]) ConcurrentModel[T, Data] {
	data := &ConcurrentData[T, Data]{}
	return data.new(cc)
}
