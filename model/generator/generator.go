package generatorModel

import (
	lockModel "kiteu/model/lock"
	"runtime"
)

type Context struct {
	arg  chan any
	ret  chan any
	err  any
	done bool
	lockModel.Lock
}
type GeneratorModel interface {
	Next(any) *Ret
	Return(any) *Ret
	Throw(any) *Ret
}

type Ret struct {
	Value any
	Done  bool
}
type JSGenerator struct {
	*Context
}

func (c *Context) Yield(v any) any {
	c.ret <- v
	receivedArg := <-c.arg
	if c.done {
		runtime.Goexit()
	}
	if c.err != nil {
		panic(c.err)
	}
	return receivedArg
}

func (c *Context) YieldGen(gen GeneratorModel) any {
	curRet := gen.Next(nil)
	for !curRet.Done {
		c.ret <- curRet.Value
		curRet = gen.Next(<-c.arg)
	}
	return curRet.Value

}

func (g *JSGenerator) nextStep(v any) any {
	g.arg <- v
	return <-g.ret
}

func (g *JSGenerator) Next(v any) (ret *Ret) {
	g.Do(func() {
		if g.done {
			ret = &Ret{Done: g.done}
			return
		}
		ret = &Ret{Value: g.nextStep(v), Done: g.done}
	})
	return ret
}

func (g *JSGenerator) Return(v any) (ret *Ret) {
	g.Do(func() {
		if !g.done {
			g.done = true
			close(g.arg)
			close(g.ret)
		}
		ret = &Ret{Done: g.done, Value: v}
	})
	return ret
}

func (g *JSGenerator) Throw(err any) (ret *Ret) {
	g.Do(func() {
		if g.done {
			panic(err)
		}
		g.err = err
		ret = &Ret{Value: g.nextStep(nil), Done: g.done}
		g.err = nil
	})
	return ret
}

func ToGenerator[T any](arr []T) GeneratorModel {
	return Generator(func(ctx *Context, _ ...any) any {
		for _, v := range arr {
			ctx.Yield(v)
		}
		return nil
	})
}

func Generator(f func(*Context, ...any) any, args ...any) GeneratorModel {
	ctx := &Context{arg: make(chan any), ret: make(chan any)}
	go func() {
		<-ctx.arg
		ret := f(ctx, args...)
		ctx.done = true
		ctx.ret <- ret
	}()
	return &JSGenerator{Context: ctx}
}
