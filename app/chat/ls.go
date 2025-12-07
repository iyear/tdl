package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
	"github.com/mattn/go-runewidth"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/texpr"
)

//go:generate go-enum --names --values --flag --nocase

type Dialog struct {
	ID          int64   `json:"id" comment:"ID of dialog"`
	Type        string  `json:"type" comment:"Type of dialog. Can be 'private', 'channel' or 'group'"`
	VisibleName string  `json:"visible_name,omitempty" comment:"Title of channel and group, first and last name of user. If empty, output '-'"`
	Username    string  `json:"username,omitempty" comment:"Username of dialog. If empty, output '-'"`
	Topics      []Topic `json:"topics,omitempty" comment:"Topics of dialog. If not set, output '-'"`
}

type Topic struct {
	ID    int    `json:"id" comment:"ID of topic"`
	Title string `json:"title" comment:"Title of topic"`
}

// ListOutput
// ENUM(table, json)
type ListOutput int

// External designation, different from Telegram mtproto
const (
	DialogGroup   = "group"
	DialogPrivate = "private"
	DialogChannel = "channel"
	DialogUnknown = "unknown"
)

type ListOptions struct {
	Output ListOutput
	Filter string
}

func List(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts ListOptions) error {
	log := logctx.From(ctx)

	// align output
	runewidth.EastAsianWidth = false
	runewidth.DefaultCondition.EastAsianWidth = false

	// output available fields
	if opts.Filter == "-" {
		fg := texpr.NewFieldsGetter(nil)
		fields, err := fg.Walk(&Dialog{})
		if err != nil {
			return fmt.Errorf("failed to walk fields: %w", err)
		}

		fmt.Print(fg.Sprint(fields, true))
		return nil
	}
	// compile filter
	filter, err := expr.Compile(opts.Filter, expr.AsBool())
	if err != nil {
		return fmt.Errorf("failed to compile filter: %w", err)
	}

	// Manually iterate through dialogs to handle errors gracefully
	// This allows us to skip problematic dialogs (deleted/inaccessible channels)
	// rather than failing completely when ExtractPeer fails
	dialogs, skipped := fetchDialogsWithErrorHandling(ctx, c.API())
	if skipped > 0 {
		log.Warn("skipped problematic dialogs during iteration",
			zap.Int("skipped", skipped),
			zap.Int("fetched", len(dialogs)))
	}

	blocked, err := tutil.GetBlockedDialogs(ctx, c.API())
	if err != nil {
		return err
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
	result := make([]*Dialog, 0, len(dialogs))
	for _, d := range dialogs {
		id := tutil.GetInputPeerID(d.Peer)

		// we can update our access hash state if there is any new peer.
		if err = applyPeers(ctx, manager, d.Entities, id); err != nil {
			log.Warn("failed to apply peer updates", zap.Int64("id", id), zap.Error(err))
		}

		// filter blocked peers
		if _, ok := blocked[id]; ok {
			continue
		}

		var r *Dialog
		switch t := d.Peer.(type) {
		case *tg.InputPeerUser:
			r = processUser(t.UserID, d.Entities)
		case *tg.InputPeerChannel:
			r = processChannel(ctx, c.API(), t.ChannelID, d.Entities)
		case *tg.InputPeerChat:
			r = processChat(t.ChatID, d.Entities)
		}

		// skip unsupported types
		if r == nil {
			continue
		}

		// filter
		b, err := texpr.Run(filter, r)
		if err != nil {
			return fmt.Errorf("failed to run filter: %w", err)
		}
		if !b.(bool) {
			continue
		}

		result = append(result, r)
	}

	switch opts.Output {
	case ListOutputTable:
		printTable(result)
	case ListOutputJson:
		bytes, err := json.MarshalIndent(result, "", "\t")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}

		fmt.Println(string(bytes))
	default:
		return fmt.Errorf("unknown output: %s", opts.Output)
	}

	return nil
}

func printTable(result []*Dialog) {
	fmt.Printf("%s %s %s %s %s\n",
		trunc("ID", 10),
		trunc("Type", 8),
		trunc("VisibleName", 20),
		trunc("Username", 20),
		"Topics")

	for _, r := range result {
		fmt.Printf("%s %s %s %s %s\n",
			trunc(strconv.FormatInt(r.ID, 10), 10),
			trunc(r.Type, 8),
			trunc(r.VisibleName, 20),
			trunc(r.Username, 20),
			topicsString(r.Topics))
	}
}

func trunc(s string, len int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = "-"
	}

	return runewidth.FillRight(runewidth.Truncate(s, len, "..."), len)
}

func topicsString(topics []Topic) string {
	if len(topics) == 0 {
		return "-"
	}

	s := make([]string, 0, len(topics))
	for _, t := range topics {
		s = append(s, fmt.Sprintf("%d: %s", t.ID, t.Title))
	}

	return strings.Join(s, ", ")
}

func processUser(id int64, entities peer.Entities) *Dialog {
	u, ok := entities.User(id)
	if !ok {
		return nil
	}

	return &Dialog{
		ID:          u.ID,
		VisibleName: visibleName(u.FirstName, u.LastName),
		Username:    u.Username,
		Type:        DialogPrivate,
		Topics:      nil,
	}
}

func processChannel(ctx context.Context, api *tg.Client, id int64, entities peer.Entities) *Dialog {
	c, ok := entities.Channel(id)
	if !ok {
		return nil
	}

	d := &Dialog{
		ID:          c.ID,
		VisibleName: c.Title,
		Username:    c.Username,
	}

	// channel type
	switch {
	case c.Broadcast:
		d.Type = DialogChannel
	case c.Megagroup, c.Gigagroup:
		d.Type = DialogGroup
	default:
		d.Type = DialogUnknown
	}

	if c.Forum {
		topics, err := fetchTopics(ctx, api, c.AsInput())
		if err != nil {
			logctx.From(ctx).Error("failed to fetch topics",
				zap.Int64("channel_id", c.ID),
				zap.String("channel_username", c.Username),
				zap.Error(err))
			return nil
		}

		d.Topics = topics
	}

	return d
}

// fetchTopics https://github.com/telegramdesktop/tdesktop/blob/4047f1733decd5edf96d125589f128758b68d922/Telegram/SourceFiles/data/data_forum.cpp#L135
func fetchTopics(ctx context.Context, api *tg.Client, c tg.InputChannelClass) ([]Topic, error) {
	log := logctx.From(ctx)
	res := make([]Topic, 0)
	limit := 100 // why can't we use 500 like tdesktop?
	offsetTopic, offsetID, offsetDate := 0, 0, 0
	lastOffsetTopic := -1 // Track the last offsetTopic to detect infinite loops

	// Track seen offsetTopics to detect cycles
	seenOffsets := make(map[int]bool)

	for {
		// Detect infinite loop: if offsetTopic hasn't changed or we've seen it before
		if offsetTopic == lastOffsetTopic && lastOffsetTopic != -1 {
			log.Warn("pagination stuck (same offset), breaking loop",
				zap.Int("offset_topic", offsetTopic))
			break
		}
		if seenOffsets[offsetTopic] {
			log.Warn("pagination cycle detected, breaking loop",
				zap.Int("offset_topic", offsetTopic))
			break
		}
		seenOffsets[offsetTopic] = true
		lastOffsetTopic = offsetTopic

		req := &tg.ChannelsGetForumTopicsRequest{
			Channel:     c,
			Limit:       limit,
			OffsetTopic: offsetTopic,
			OffsetID:    offsetID,
			OffsetDate:  offsetDate,
		}

		topics, err := api.ChannelsGetForumTopics(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "get forum topics")
		}

		// If no topics returned, we're done
		if len(topics.Topics) == 0 {
			break
		}

		for _, tp := range topics.Topics {
			if t, ok := tp.(*tg.ForumTopic); ok {
				res = append(res, Topic{
					ID:    t.ID,
					Title: t.Title,
				})

				offsetTopic = t.ID
			}
		}

		// Safety break if we've collected all topics
		if len(res) >= topics.Count {
			break
		}

		// last page
		if len(topics.Topics) < limit {
			break
		}

		// Update offset using last message if available
		// Use a local variable for length to be absolutely safe against index out of range
		msgCount := len(topics.Messages)
		if msgCount > 0 {
			if lastMsg, ok := topics.Messages[msgCount-1].AsNotEmpty(); ok {
				offsetID, offsetDate = lastMsg.GetID(), lastMsg.GetDate()
			} else {
				log.Debug("no valid message for offset, relying on offsetTopic only",
					zap.Int("offset_topic", offsetTopic))
			}
		} else {
			log.Debug("no messages in topics response, relying on offsetTopic only",
				zap.Int("offset_topic", offsetTopic),
				zap.Int("topics_count", len(topics.Topics)))
		}
	}

	return res, nil
}

func processChat(id int64, entities peer.Entities) *Dialog {
	c, ok := entities.Chat(id)
	if !ok {
		return nil
	}

	return &Dialog{
		ID:          c.ID,
		VisibleName: c.Title,
		Username:    "-",
		Type:        DialogGroup,
		Topics:      nil,
	}
}

func visibleName(first, last string) string {
	if first == "" && last == "" {
		return ""
	}

	if first == "" {
		return last
	}

	if last == "" {
		return first
	}

	return first + " " + last
}

func applyPeers(ctx context.Context, manager *peers.Manager, entities peer.Entities, id int64) error {
	users := make([]tg.UserClass, 0, 1)
	if user, ok := entities.User(id); ok {
		users = append(users, user)
	}

	chats := make([]tg.ChatClass, 0, 1)
	if chat, ok := entities.Chat(id); ok {
		chats = append(chats, chat)
	}
	if channel, ok := entities.Channel(id); ok {
		chats = append(chats, channel)
	}

	return manager.Apply(ctx, users, chats)
}

// fetchDialogsWithErrorHandling manually iterates through dialogs using the raw Telegram API
// to gracefully handle errors from problematic dialogs (deleted/inaccessible channels).
// Instead of failing completely when ExtractPeer fails in gotd's iterator, it logs errors
// and continues, skipping bad dialogs.
func fetchDialogsWithErrorHandling(ctx context.Context, api *tg.Client) ([]dialogs.Elem, int) {
	log := logctx.From(ctx)
	const batchSize = 100
	var (
		allElems   []dialogs.Elem
		skipped    int
		offsetID   int
		offsetDate int
		offsetPeer tg.InputPeerClass = &tg.InputPeerEmpty{}
		seen                         = make(map[int64]bool) // Track seen dialog IDs to prevent duplicates
	)

	for {
		// Fetch a batch of dialogs using raw API
		result, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetDate: offsetDate,
			OffsetID:   offsetID,
			OffsetPeer: offsetPeer,
			Limit:      batchSize,
		})
		if err != nil {
			log.Error("failed to fetch dialog batch", zap.Error(err))
			break
		}

		var (
			dialogsSlice []tg.DialogClass
			messages     []tg.MessageClass
			users        []tg.UserClass
			chats        []tg.ChatClass
		)

		switch d := result.(type) {
		case *tg.MessagesDialogs:
			dialogsSlice = d.Dialogs
			messages = d.Messages
			users = d.Users
			chats = d.Chats
		case *tg.MessagesDialogsSlice:
			dialogsSlice = d.Dialogs
			messages = d.Messages
			users = d.Users
			chats = d.Chats
		case *tg.MessagesDialogsNotModified:
			// No more dialogs
			return allElems, skipped
		default:
			log.Error("unexpected dialog type", zap.String("type", fmt.Sprintf("%T", result)))
			return allElems, skipped
		}

		if len(dialogsSlice) == 0 {
			break
		}

		// Build entities map for this batch
		// Convert slices to maps as required by peer.NewEntities
		userMap := make(map[int64]*tg.User)
		for _, u := range users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		chatMap := make(map[int64]*tg.Chat)
		channelMap := make(map[int64]*tg.Channel)
		for _, c := range chats {
			switch chat := c.(type) {
			case *tg.Chat:
				chatMap[chat.ID] = chat
			case *tg.Channel:
				channelMap[chat.ID] = chat
			}
		}

		entities := peer.NewEntities(userMap, chatMap, channelMap)

		// Build message map for quick lookup by ID
		messageMap := make(map[int]tg.NotEmptyMessage)
		for _, msg := range messages {
			switch m := msg.(type) {
			case *tg.Message:
				messageMap[m.ID] = m
			case *tg.MessageService:
				messageMap[m.ID] = m
			}
		}

		// Process each dialog in this batch
		for _, d := range dialogsSlice {
			dialog, ok := d.(*tg.Dialog)
			if !ok {
				continue
			}

			// Find the peer ID for logging purposes
			var peerID int64
			switch p := dialog.Peer.(type) {
			case *tg.PeerUser:
				peerID = p.UserID
			case *tg.PeerChat:
				peerID = p.ChatID
			case *tg.PeerChannel:
				peerID = p.ChannelID
			default:
				log.Error("unknown peer type", zap.String("type", fmt.Sprintf("%T", p)))
				skipped++
				continue
			}

			// Skip if we've already seen this dialog (deduplication)
			if seen[peerID] {
				continue
			}
			seen[peerID] = true

			// Try to extract the peer - THIS IS WHERE THE ORIGINAL ERROR OCCURS
			// In gotd's query/dialogs iterator, it calls ExtractPeer without error handling,
			// causing a panic when a channel doesn't exist in entities.
			// We catch it here and skip the problematic dialog instead.
			// See: https://github.com/iyear/tdl/issues/713
			inputPeer, err := entities.ExtractPeer(dialog.Peer)
			if err != nil {
				// This dialog references a channel/chat that doesn't exist in entities
				// (likely deleted, user was banned, or channel is inaccessible).
				// Log and skip it instead of failing.
				log.Warn("skipping dialog with missing peer",
					zap.Int64("peer_id", peerID),
					zap.String("peer_type", fmt.Sprintf("%T", dialog.Peer)),
					zap.Error(err))
				skipped++
				continue
			}

			// Get the last message for this dialog from message map
			lastMsg := messageMap[dialog.TopMessage]

			// Successfully processed this dialog
			allElems = append(allElems, dialogs.Elem{
				Peer:     inputPeer,
				Entities: entities,
				Dialog:   dialog,
				Last:     lastMsg,
			})
		}

		// Update offset for next batch using the last dialog in dialogsSlice
		// (regardless of whether it was successfully processed or skipped)
		if len(dialogsSlice) > 0 {
			lastDialog, ok := dialogsSlice[len(dialogsSlice)-1].(*tg.Dialog)
			if ok {
				// Get the message date from message map
				var msgDate int
				if lastMsg, found := messageMap[lastDialog.TopMessage]; found {
					msgDate = lastMsg.GetDate()
				}

				offsetDate = msgDate
				offsetID = lastDialog.TopMessage

				// Set offset peer based on dialog peer type
				// Try to get access hash from entities if available
				switch peerType := lastDialog.Peer.(type) {
				case *tg.PeerUser:
					if user, ok := entities.User(peerType.UserID); ok {
						offsetPeer = &tg.InputPeerUser{
							UserID:     peerType.UserID,
							AccessHash: user.AccessHash,
						}
					} else {
						// Can't continue pagination without access hash
						log.Error("failed to get user for offset, stopping pagination",
							zap.Int64("user_id", peerType.UserID))
						color.Red("Error: failed to get user for offset, stopping pagination. User ID: %d", peerType.UserID)
						return allElems, skipped
					}
				case *tg.PeerChat:
					offsetPeer = &tg.InputPeerChat{ChatID: peerType.ChatID}
				case *tg.PeerChannel:
					if channel, ok := entities.Channel(peerType.ChannelID); ok {
						offsetPeer = &tg.InputPeerChannel{
							ChannelID:  peerType.ChannelID,
							AccessHash: channel.AccessHash,
						}
					} else {
						// Can't continue pagination without access hash
						log.Error("failed to get channel for offset, stopping pagination",
							zap.Int64("channel_id", peerType.ChannelID))
						color.Red("Error: failed to get channel for offset, stopping pagination. Channel ID: %d", peerType.ChannelID)
						return allElems, skipped
					}
				}
			}
		}

		// Check if we've fetched all dialogs
		// Continue fetching if we got a full batch (there might be more)
		if len(dialogsSlice) < batchSize {
			// Got less than requested, we've reached the end
			break
		}
	}

	return allElems, skipped
}
