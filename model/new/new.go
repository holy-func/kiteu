package newModel

import (
	"kiteu/util"
	"sync"
)

type Init struct {
	once sync.Once
}

func (model *Init) InitCall(init util.Callback) {
	model.once.Do(func() { util.SafeCallback(init) })
}
