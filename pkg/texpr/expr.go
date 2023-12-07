package texpr

import (
	"sync"

	"github.com/antonmedv/expr/vm"
)

var vmPool = sync.Pool{New: func() any {
	return &vm.VM{}
}}

// TODO(iyear): function helpers

func Run(program *vm.Program, env any) (any, error) {
	v := vmPool.Get().(*vm.VM)
	defer vmPool.Put(v)

	return v.Run(program, env)
}
