package takeout

import (
	"context"
	"errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type takeout struct {
	id int64
}

type nopDecoder struct {
	bin.Encoder
}

func (n nopDecoder) Decode(_ *bin.Buffer) error {
	return errors.New("bin.Decoder is not implemented")
}

func (t takeout) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		return next.Invoke(ctx, &tg.InvokeWithTakeoutRequest{
			TakeoutID: t.id,
			Query:     nopDecoder{input},
		}, output)
	}
}

func Middleware(id int64) telegram.Middleware {
	return takeout{id: id}
}
