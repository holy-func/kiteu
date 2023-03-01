package design

type Strategy[T any] struct {
	Name    string
	Match   func(T) bool
	Execute func(T)
}
type Dispatch[T any] interface {
	Execute(T)
}
type CommonDispatch[T any] struct {
	Strategys   []*Strategy[T]
	MatchOne    bool
	DefaultExec func(T)
}

func (cd *CommonDispatch[T]) Execute(v T) {
	matched := false
	for _, strategy := range cd.Strategys {
		if strategy.Match(v) {
			strategy.Execute(v)
			matched = true
			if cd.MatchOne {
				return
			}
		}
	}
	if !matched && cd.DefaultExec != nil {
		cd.DefaultExec(v)
	}
}
