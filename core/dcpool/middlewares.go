package dcpool

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func chainMiddlewares(invoker tg.Invoker, chain ...telegram.Middleware) tg.Invoker {
	if len(chain) == 0 {
		return invoker
	}
	for i := len(chain) - 1; i >= 0; i-- {
		invoker = chain[i].Handle(invoker)
	}

	return invoker
}
