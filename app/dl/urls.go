package dl

import (
	"context"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/utils"
)

func parseURLs(ctx context.Context, client *tg.Client, urls []string) ([]*dialog, error) {
	manager := peers.Options{}.Build(client)
	msgMap := make(map[int64]*dialog)

	for _, u := range urls {
		ch, msgid, err := utils.Telegram.ParseChannelMsgLink(ctx, manager, u)
		if err != nil {
			return nil, err
		}

		// init map value
		if _, ok := msgMap[ch.ChannelID]; !ok {
			msgMap[ch.ChannelID] = &dialog{peer: &tg.InputPeerChannel{ChannelID: ch.ChannelID, AccessHash: ch.AccessHash}, msgs: []int{}}
		}

		msgMap[ch.ChannelID].msgs = append(msgMap[ch.ChannelID].msgs, msgid)
	}

	// cap is at least len of map
	msgs := make([]*dialog, 0, len(msgMap))
	for _, m := range msgMap {
		msgs = append(msgs, m)
	}

	return msgs, nil
}
