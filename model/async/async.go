package asyncModel

import promiseModel "kiteu/model/promise"

type GoAsync[T any] func() T

func Async[T any](task GoAsync[T]) promiseModel.GoPromise[T] {
	return promiseModel.Promise[T](func(resolve, _ func(any)) {
		go resolve(task())
	})
}
func Await[T any](promise promiseModel.GoPromise[T]) T {
	return promise.Await()
}
