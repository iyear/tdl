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
		client := pool.Default(ctx)

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

			// Check for grouped messages
			singleMsg, err := tutil.GetSingleMessage(ctx, client, ch.InputPeer(), msgid)
			if err != nil {
				logctx.From(ctx).Warn("GetSingleMessage failed", zap.Error(err))
			} else if _, ok := singleMsg.GetGroupedID(); ok {
				groupedMessages, err := tutil.GetGroupedMessages(ctx, client, ch.InputPeer(), singleMsg)
				if err != nil {
					logctx.From(ctx).Warn("GetGroupedMessages failed", zap.Error(err))
				} else {
					// Add all message IDs from the group
					msgIDs := msgMap[ch.ID()].Messages
					for _, gm := range groupedMessages {
						isNew := true
						for _, existingID := range msgIDs {
							if existingID == gm.ID {
								isNew = false
								break
							}
						}
						if isNew {
							msgMap[ch.ID()].Messages = append(msgMap[ch.ID()].Messages, gm.ID)
						}
					}
				}
			}
		}

		dialogList := make([]*Dialog, 0, len(msgMap))
		for _, dialog := range msgMap {
			dialogList = append(dialogList, dialog)
		}

		return dialogList, nil
	}
}
