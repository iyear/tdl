package tmessage

import (
	"context"

	"github.com/gotd/td/telegram/peers"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/util/tutil"
)

func FromURL(ctx context.Context, pool dcpool.Pool, kvd storage.Storage, urls []string) ParseSource {
	return func() ([]*Dialog, error) {
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.
			Build(pool.Default(ctx))
		msgMap := make(map[int64]*Dialog)

		for _, u := range urls {
			ch, msgid, err := tutil.ParseMessageLink(ctx, manager, u)
			if err != nil {
				return nil, err
			}
			logctx.From(ctx).Debug("Parse URL",
				zap.String("url", u),
				zap.Int64("peer_id", ch.ID()),
				zap.String("peer_name", ch.VisibleName()),
				zap.Int("msg", msgid))

			// init map value
			if _, ok := msgMap[ch.ID()]; !ok {
				msgMap[ch.ID()] = &Dialog{Peer: ch.InputPeer(), Messages: []int{}}
			}

			msgMap[ch.ID()].Messages = append(msgMap[ch.ID()].Messages, msgid)
		}

		// cap is at least len of map
		msgs := make([]*Dialog, 0, len(msgMap))
		for _, m := range msgMap {
			msgs = append(msgs, m)
		}

		return msgs, nil
	}
}
