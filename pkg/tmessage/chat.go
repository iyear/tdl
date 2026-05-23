package tmessage

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/util/tutil"
)

// FromChat creates a ParseSource that collects all media message IDs from a chat
// by iterating the chat history via the Telegram API. This avoids the need to
// pass individual message links or export JSON files.
//
// chat: chat ID (numeric) or username
// topic: topic root message ID (0 = no topic, iterate full history)
// msgStart: minimum message ID to include (0 = no lower bound)
// msgEnd: maximum message ID to include (0 = no upper bound)
func FromChat(ctx context.Context, pool dcpool.Pool, kvd storage.Storage, chat string, topic int, msgStart int, msgEnd int) ParseSource {
	return func() ([]*Dialog, error) {
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.
			Build(pool.Default(ctx))

		// Normalize chat ID: handle -100XXXXXXXXXX format
		normalizedChat := normalizeChatID(chat)

		peer, err := tutil.GetInputPeer(ctx, manager, normalizedChat)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve chat '%s'", chat)
		}

		logctx.From(ctx).Info("Resolved chat for download",
			zap.Int64("peer_id", peer.ID()),
			zap.String("peer_name", peer.VisibleName()),
			zap.Int("topic", topic),
			zap.Int("msg_start", msgStart),
			zap.Int("msg_end", msgEnd))

		color.Cyan("Collecting messages from '%s' (ID: %d)...", peer.VisibleName(), peer.ID())

		// Build the appropriate query (topic or full history)
		var q messages.Query
		if topic > 0 {
			q = query.NewQuery(pool.Default(ctx)).Messages().GetReplies(peer.InputPeer()).MsgID(topic)
		} else {
			q = query.NewQuery(pool.Default(ctx)).Messages().GetHistory(peer.InputPeer())
		}

		iter := messages.NewIterator(q, 100)

		// If msgEnd is set, start iterating from that point
		if msgEnd > 0 {
			iter = iter.OffsetID(msgEnd + 1)
		}

		msgIDs := make([]int, 0)
		textMsgs := make([]TextMsg, 0)
		for iter.Next(ctx) {
			msg := iter.Value()

			m, ok := msg.Msg.(*tg.Message)
			if !ok {
				continue
			}

			// Stop if we've gone past the start boundary
			if msgStart > 0 && m.ID < msgStart {
				break
			}

			// Collect media messages for download
			if hasMedia(m) {
				msgIDs = append(msgIDs, m.ID)
			}

			// Collect text messages for HTML export (including captions on media messages)
			if m.Message != "" {
				tm := TextMsg{
					ID:       m.ID,
					Date:     m.Date,
					Text:     m.Message,
					Entities: m.Entities,
				}

				// Extract reply info
				if replyTo, ok := m.GetReplyTo(); ok {
					if rh, ok := replyTo.(*tg.MessageReplyHeader); ok {
						tm.ReplyToMsgID = rh.ReplyToMsgID
					}
				}

				// Try to extract sender name from the message's FromID
				if fromID, ok := m.GetFromID(); ok {
					tm.FromName = extractSenderName(ctx, manager, fromID)
				}

				textMsgs = append(textMsgs, tm)
			}
		}

		if err := iter.Err(); err != nil {
			return nil, errors.Wrap(err, "iterate chat messages")
		}

		if len(msgIDs) == 0 && len(textMsgs) == 0 {
			return nil, fmt.Errorf("no messages found in chat '%s' (ID: %d)", peer.VisibleName(), peer.ID())
		}

		if len(msgIDs) > 0 {
			color.Green("Found %d media messages to download", len(msgIDs))
		}
		if len(textMsgs) > 0 {
			color.Green("Found %d text messages for HTML export", len(textMsgs))
		}

		return []*Dialog{{
			Peer:         peer.InputPeer(),
			Messages:     msgIDs,
			TextMessages: textMsgs,
		}}, nil
	}
}

// extractSenderName tries to resolve a PeerClass into a display name.
func extractSenderName(ctx context.Context, manager *peers.Manager, fromID tg.PeerClass) string {
	switch f := fromID.(type) {
	case *tg.PeerUser:
		if p, err := manager.ResolveUserID(ctx, f.UserID); err == nil {
			return p.VisibleName()
		}
	case *tg.PeerChannel:
		if p, err := manager.ResolveChannelID(ctx, f.ChannelID); err == nil {
			return p.VisibleName()
		}
	case *tg.PeerChat:
		if p, err := manager.ResolveChatID(ctx, f.ChatID); err == nil {
			return p.VisibleName()
		}
	}
	return ""
}

// hasMedia checks if a message contains downloadable media (document or photo).
func hasMedia(m *tg.Message) bool {
	md, ok := m.GetMedia()
	if !ok {
		return false
	}

	switch md.(type) {
	case *tg.MessageMediaDocument, *tg.MessageMediaPhoto:
		return true
	default:
		return false
	}
}

// normalizeChatID handles the common Telegram marked channel ID format.
// Users often encounter IDs like -1001931890116 (from bot APIs), but the
// internal Telegram channel ID is 1931890116. This strips the -100 prefix.
func normalizeChatID(chat string) string {
	if strings.HasPrefix(chat, "-100") && len(chat) > 4 {
		return chat[4:]
	}
	return chat
}
