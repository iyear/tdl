package tmessage

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tmedia"
	"github.com/iyear/tdl/core/util/tutil"
)

// FromPaidHistory scans complete chat histories and selects paid media that is
// unlocked for the current account. Locked previews are not added to the
// download queue.
func FromPaidHistory(ctx context.Context, pool dcpool.Pool, kvd storage.Storage, chats []string) ParseSource {
	return func() ([]*Dialog, error) {
		if len(chats) == 0 {
			return nil, nil
		}

		client := pool.Default(ctx)
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(client)
		dialogs := make([]*Dialog, 0, len(chats))
		seen := make(map[string]struct{}, len(chats))

		for _, chat := range chats {
			peer, err := tutil.GetInputPeer(ctx, manager, chat)
			if err != nil {
				return nil, errors.Wrapf(err, "resolve paid media chat %q", chat)
			}
			peerKey := fmt.Sprintf("%T:%d", peer.InputPeer(), peer.ID())
			if _, ok := seen[peerKey]; ok {
				continue
			}
			seen[peerKey] = struct{}{}

			dialog := &Dialog{
				Peer:        peer.InputPeer(),
				Messages:    make([]int, 0),
				MediaCounts: make(map[int]int),
			}
			iter := messages.NewIterator(
				query.NewQuery(client).Messages().GetHistory(peer.InputPeer()),
				100,
			)
			for iter.Next(ctx) {
				message, ok := iter.Value().Msg.(*tg.Message)
				if !ok {
					continue
				}

				count := unlockedPaidMediaCount(message)
				if count == 0 {
					continue
				}
				dialog.Messages = append(dialog.Messages, message.ID)
				dialog.MediaCounts[message.ID] = count
			}
			if err := iter.Err(); err != nil {
				return nil, errors.Wrapf(err, "scan paid media chat %q", chat)
			}

			if len(dialog.Messages) > 0 {
				dialogs = append(dialogs, dialog)
			}
		}

		if len(dialogs) == 0 {
			return nil, errors.New("no unlocked paid media found")
		}
		return dialogs, nil
	}
}

func unlockedPaidMediaCount(message *tg.Message) int {
	media, ok := message.GetMedia()
	if !ok {
		return 0
	}
	paid, ok := media.(*tg.MessageMediaPaidMedia)
	if !ok {
		return 0
	}

	count := 0
	for _, item := range tmedia.GetPaidMedia(paid) {
		if item != nil {
			count++
		}
	}
	return count
}
