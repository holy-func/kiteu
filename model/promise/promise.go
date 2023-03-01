package promiseModel

import (
	"fmt"
	"kiteu/design"
	concurrentModel "kiteu/model/concurrent"
	newModel "kiteu/model/new"
	"kiteu/util"
	"reflect"
	"sync"
)

// todo 构造任务队列 或 reject错误异步抛出 若被处理则取消抛出

type Executor[T any] func(resolve, reject func(any))
type Handler[T any] func(v T) any
type GoPromise[T any] interface {
	Then(Handler[T], Handler[any]) GoPromise[any]
	Catch(Handler[any]) GoPromise[any]
	Finally(func()) GoPromise[any]
	Await() T
	String() string
}
type promiseState string

const CanNotResolveSelf = "can not resolve itself"

const (
	pending   promiseState = "pending"
	fulfilled promiseState = "fulfilled"
	rejected  promiseState = "rejected"
)

type Thenable interface {
	Then(func(any), func(any))
}

type resolvePayload struct {
	val         reflect.Value
	promise     reflect.Value
	onFulfilled Handler[any]
	onRejected  Handler[any]
}

var resolveStrategy *design.CommonDispatch[*resolvePayload]

func getResolveStrategy() *design.CommonDispatch[*resolvePayload] {
	if resolveStrategy != nil {
		return resolveStrategy
	}
	resolveStrategy = &design.CommonDispatch[*resolvePayload]{
		MatchOne: true,
		Strategys: []*design.Strategy[*resolvePayload]{
			{
				Name: "thenable",
				Match: func(rp *resolvePayload) bool {
					return util.IsInterface[Thenable](rp.val)
				},
				Execute: func(rp *resolvePayload) {
					then := util.GetUnderlying[Thenable](rp.val)
					var thenVal any
					util.TryCatch(&util.TryCatchBody{
						Try: func() {
							Promise[any](func(resolve, reject func(any)) {
								Promise[any](func(innerResolve, innerReject func(any)) {
									resolveInner := func(v any) {
										thenVal = v
										innerResolve(nil)
									}
									rejectBoth := func(r any) {
										innerReject(r)
										reject(r)
									}
									then.Then(resolveInner, rejectBoth)
								}).Finally(func() {
									if util.IsInterface[Thenable](thenVal) && util.GetUnderlying[Thenable](thenVal) == then {
										reject(CanNotResolveSelf)
										return
									}
									resolve(thenVal)
								})
							}).
								Then(
									rp.onFulfilled,
									rp.onRejected,
								)
						},
						Catch: func(err any) {
							rp.onRejected(err)
						},
					})

				},
			},
			{
				Name: "promise",
				Match: func(rp *resolvePayload) bool {
					return util.IsType[promiseBody[any]](rp.val)
				},
				Execute: func(rp *resolvePayload) {
					if util.IdempotentReflectValueOf(rp.promise) == util.IdempotentReflectValueOf(rp.val) {
						rp.onRejected(CanNotResolveSelf)
						return
					}
					onFulfilled := reflect.MakeFunc(rp.val.MethodByName("Then").Type().In(0), func(args []reflect.Value) (results []reflect.Value) {
						return []reflect.Value{reflect.ValueOf(rp.onFulfilled(args[0].Interface()))}
					})
					rp.val.MethodByName("Then").Call([]reflect.Value{
						onFulfilled, reflect.ValueOf(rp.onRejected),
					})
				},
			},
		},
		DefaultExec: func(rp *resolvePayload) {
			rp.onFulfilled(rp.val.Interface())
		},
	}
	return resolveStrategy
}

func (p promiseState) fulfilled() bool {
	return p == fulfilled
}
func (p promiseState) rejected() bool {
	return p == rejected
}

type msg[T any] struct {
	resolve     func(any)
	reject      func(any)
	onFulfilled Handler[T]
	onRejected  Handler[any]
}

func (m *msg[T]) Execute(state promiseState, val any) {
	go util.TryCatch(&util.TryCatchBody{
		Try: func() {
			if state.fulfilled() {
				if util.IsSafeFunc(m.onFulfilled) {
					m.resolve(m.onFulfilled(val.(T)))
					return
				}
				m.resolve(val)
			} else if state.rejected() {
				if util.IsSafeFunc(m.onRejected) {
					m.resolve(m.onRejected(val))
					return
				}
				m.reject(val)
			}
		},
		Catch: func(err any) {
			m.reject(err)
		},
	})
}

type promiseBody[T any] struct {
	lock  sync.Once
	state promiseState
	val   any
	model concurrentModel.ConcurrentModel[[]*msg[T], *msg[T]]
	newModel.Init
}

func Promise[T any](executor Executor[T]) GoPromise[T] {
	p := &promiseBody[T]{}
	return p.new(executor)
}

func Resolve[T any](v any) GoPromise[T] {
	return Promise[T](func(resolve, _ func(any)) {
		resolve(v)
	})
}

func (p *promiseBody[T]) new(executor Executor[T]) GoPromise[T] {
	p.InitCall(func() {
		p.state = pending
		model := concurrentModel.CreateConcurrent(concurrentModel.ConcurrentSlice[*msg[T]](2))
		p.model = model
		util.TryCatch(&util.TryCatchBody{
			Try: func() {
				executor(p.resolve, p.reject)
			},
			Catch: func(err any) {
				p.reject(err)
			},
		})
	})
	return p
}

func (p *promiseBody[T]) String() string {
	if p.state.rejected() {
		return fmt.Sprintf("Promise { <%v> %v }", p.state, p.val)
	} else if p.state.fulfilled() {
		return fmt.Sprintf("Promise { %v }", p.val)
	} else {
		return fmt.Sprintf("Promise { <%v> }", p.state)
	}
}

func (p *promiseBody[T]) onFulfilled(val any) any {
	p.done(fulfilled, val)
	return nil
}
func (p *promiseBody[T]) onRejected(reason any) any {
	p.done(rejected, reason)
	return nil
}

func (p *promiseBody[T]) done(state promiseState, val any) {
	p.lock.Do(func() {
		p.state = state
		p.val = val
		p.model.Seal()
		msgs:=p.model.GetData()
		for _, m := range msgs {
			m.Execute(state, val)
		}
	})
}
func (p *promiseBody[T]) Await() T {
	p.model.Wait()
	return p.val.(T)
}
func (p *promiseBody[T]) resolve(v any) {
	getResolveStrategy().Execute(&resolvePayload{
		val:         reflect.ValueOf(v),
		promise:     reflect.ValueOf(p),
		onFulfilled: p.onFulfilled,
		onRejected:  p.onRejected,
	})
}
func (p *promiseBody[T]) reject(v any) {
	p.onRejected(v)
}
func (p *promiseBody[T]) Then(onFulfilled Handler[T], onRejected Handler[any]) GoPromise[any] {
	ret := Promise[any](func(resolve, reject func(any)) {
		m := &msg[T]{resolve: resolve, reject: reject, onFulfilled: onFulfilled, onRejected: onRejected}
		if p.model.IsSealed() {
			m.Execute(p.state, p.val)
		} else {
			p.model.AddData(m)
		}
	})
	return ret

}
func (p *promiseBody[T]) Catch(catch Handler[any]) GoPromise[any] {
	return p.Then(nil, catch)
}
func (p *promiseBody[T]) Finally(callback func()) GoPromise[any] {
	return p.Then(func(v T) any {
		callback()
		return v
	}, func(v any) any {
		callback()
		return v
	})
}
