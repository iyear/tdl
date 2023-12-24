package retry

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/logger"
)

var internalErrors = []string{
	"Timedout", // #373
	"No workers running",
	"RPC_CALL_FAIL",
	"RPC_MCGET_FAIL",
}

type retry struct {
	max    int
	errors []string
}

func (r retry) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		retries := 0

		for retries < r.max {
			if err := next.Invoke(ctx, input, output); err != nil {
				if tgerr.Is(err, r.errors...) {
					logger.From(ctx).Debug("retry middleware", zap.Int("retries", retries), zap.Error(err))
					retries++
					continue
				}
				return errors.Wrap(err, "retry middleware skip")
			}

			return nil
		}

		return fmt.Errorf("retry limit reached after %d attempts", r.max)
	}
}

// New returns middleware that retries request if it fails with one of provided errors.
func New(max int, errors ...string) telegram.Middleware {
	return retry{
		max:    max,
		errors: append(errors, internalErrors...), // #373
	}
}
