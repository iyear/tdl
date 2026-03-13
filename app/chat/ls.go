package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
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
	// rather than failing completely when ExtractPeer fails.
	// We can't use query.GetDialogs() iterator because it fails when trying to
	// extract peers for pagination offsets. We use manual pagination instead.
	var allDialogs []dialogs.Elem
	var skipped int
	var nonDialogCount int // Count of DialogFolder or other non-Dialog types

	// Accumulate entities and messages across all batches
	globalUserMap := make(map[int64]*tg.User)
	globalChatMap := make(map[int64]*tg.Chat)
	globalChannelMap := make(map[int64]*tg.Channel)
	globalMessageMap := make(map[int]tg.NotEmptyMessage)
	allRawDialogs := make([]*tg.Dialog, 0)
	seenDialogs := make(map[int64]bool) // Track seen dialog IDs to prevent duplicates

	// Pagination state
	offsetID := 0
	offsetDate := 0
	const batchSize = 100
	batchNum := 0

	for {
		// Use the raw API with pagination
		apiResult, err := c.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetDate: offsetDate,
			OffsetID:   offsetID,
			OffsetPeer: &tg.InputPeerEmpty{}, // Use empty peer to avoid extraction issues
			Limit:      batchSize,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		var (
			dialogsSlice []tg.DialogClass
			messages     []tg.MessageClass
			users        []tg.UserClass
			chats        []tg.ChatClass
			totalCount   int // Total number of dialogs available
		)

		switch d := apiResult.(type) {
		case *tg.MessagesDialogs:
			// Complete list of dialogs (all fetched in one call)
			dialogsSlice = d.Dialogs
			messages = d.Messages
			users = d.Users
			chats = d.Chats
			totalCount = len(d.Dialogs)
		case *tg.MessagesDialogsSlice:
			// Partial list (pagination needed)
			dialogsSlice = d.Dialogs
			messages = d.Messages
			users = d.Users
			chats = d.Chats
			totalCount = d.Count // Total available dialogs from API
		case *tg.MessagesDialogsNotModified:
			// No more dialogs
			log.Debug("received dialogsNotModified, ending pagination", zap.Int("batch", batchNum))
			return nil
		default:
			return fmt.Errorf("unexpected dialogs type: %T", apiResult)
		}

		if len(dialogsSlice) == 0 {
			log.Debug("no more dialogs, ending pagination", zap.Int("batch", batchNum))
			break
		}

		batchNum++
		log.Info("fetched dialog batch",
			zap.Int("batch_num", batchNum),
			zap.Int("batch_size", len(dialogsSlice)),
			zap.Int("total_count", totalCount),
			zap.Int("fetched_so_far", len(allRawDialogs)+len(dialogsSlice)),
			zap.Int("messages", len(messages)),
			zap.Int("users", len(users)),
			zap.Int("chats", len(chats)),
			zap.Int("offset_id", offsetID),
			zap.Int("offset_date", offsetDate))

		// Accumulate users into global map
		for _, u := range users {
			if user, ok := u.(*tg.User); ok {
				globalUserMap[user.ID] = user
			}
		}

		// Accumulate chats into global map
		for _, c := range chats {
			switch chat := c.(type) {
			case *tg.Chat:
				globalChatMap[chat.ID] = chat
			case *tg.Channel:
				globalChannelMap[chat.ID] = chat
			}
		}

		// Accumulate messages into global map
		for _, msg := range messages {
			if m, ok := msg.AsNotEmpty(); ok {
				globalMessageMap[m.GetID()] = m
			}
		}

		// Store raw dialogs for later processing
		var newDialogsInBatch int
		for _, d := range dialogsSlice {
			if dialog, ok := d.(*tg.Dialog); ok {
				// Get peer ID for deduplication
				var peerID int64
				switch p := dialog.Peer.(type) {
				case *tg.PeerUser:
					peerID = p.UserID
				case *tg.PeerChat:
					peerID = p.ChatID
				case *tg.PeerChannel:
					peerID = p.ChannelID
				}

				// Skip if we've already seen this dialog (deduplication)
				if seenDialogs[peerID] {
					log.Debug("skipping duplicate dialog",
						zap.Int64("peer_id", peerID),
						zap.String("peer_type", fmt.Sprintf("%T", dialog.Peer)))
					continue
				}
				seenDialogs[peerID] = true

				allRawDialogs = append(allRawDialogs, dialog)
				newDialogsInBatch++
			} else {
				// Track non-Dialog types (e.g., DialogFolder)
				nonDialogCount++
				log.Debug("skipping non-Dialog type",
					zap.String("type", fmt.Sprintf("%T", d)))
			}
		}

		// Update pagination offsets from the CURRENT BATCH's last dialog
		// This is critical - if we use allRawDialogs[last], we'll get stuck when all dialogs are duplicates
		if len(dialogsSlice) > 0 {
			// Find the last actual dialog (not DialogFolder) in this batch
			for i := len(dialogsSlice) - 1; i >= 0; i-- {
				if dialog, ok := dialogsSlice[i].(*tg.Dialog); ok {
					if msg, ok := globalMessageMap[dialog.TopMessage]; ok {
						offsetID = msg.GetID()
						offsetDate = msg.GetDate()
						log.Debug("updated pagination offsets from current batch",
							zap.Int("new_offset_id", offsetID),
							zap.Int("new_offset_date", offsetDate),
							zap.Int("dialog_count", len(allRawDialogs)),
							zap.Int("new_in_batch", newDialogsInBatch))
						break
					}
				}
			}
		}

		// Continue fetching until we have all dialogs
		// Primary stopping condition: we've fetched all dialogs according to totalCount
		if totalCount > 0 && len(allRawDialogs) >= totalCount {
			log.Info("fetched all dialogs according to total count",
				zap.Int("total_fetched", len(allRawDialogs)),
				zap.Int("total_count_from_api", totalCount))
			break
		}

		// Secondary stopping condition: if we got a full batch of duplicates, we're looping
		// This prevents infinite loops when pagination gets stuck
		if newDialogsInBatch == 0 && len(dialogsSlice) > 0 {
			log.Warn("received full batch of duplicates, pagination appears stuck",
				zap.Int("batch_size", len(dialogsSlice)),
				zap.Int("total_fetched", len(allRawDialogs)),
				zap.Int("total_count_from_api", totalCount))
			break
		}
	} // Build global entities from accumulated data
	entities := peer.NewEntities(globalUserMap, globalChatMap, globalChannelMap)

	// Now process all collected dialogs with the complete global entities
	for _, dialog := range allRawDialogs {
		// Try to extract the peer
		inputPeer, err := entities.ExtractPeer(dialog.Peer)
		if err != nil {
			// Skip dialogs with missing/invalid peers (deleted channels, etc.)
			var peerID int64
			switch p := dialog.Peer.(type) {
			case *tg.PeerUser:
				peerID = p.UserID
			case *tg.PeerChat:
				peerID = p.ChatID
			case *tg.PeerChannel:
				peerID = p.ChannelID
			}
			log.Warn("skipping dialog with invalid peer",
				zap.Int64("peer_id", peerID),
				zap.String("peer_type", fmt.Sprintf("%T", dialog.Peer)),
				zap.Error(err))
			skipped++
			continue
		}

		// Get the last message for this dialog
		lastMsg := globalMessageMap[dialog.TopMessage]

		allDialogs = append(allDialogs, dialogs.Elem{
			Peer:     inputPeer,
			Entities: entities,
			Dialog:   dialog,
			Last:     lastMsg,
		})
	}

	if skipped > 0 {
		log.Warn("skipped problematic dialogs during iteration",
			zap.Int("skipped", skipped),
			zap.Int("fetched", len(allDialogs)))
	}

	blocked, err := tutil.GetBlockedDialogs(ctx, c.API())
	if err != nil {
		return err
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
	result := make([]*Dialog, 0, len(allDialogs))

	var processedCount, nilCount, blockedCount int
	for _, d := range allDialogs {
		id := tutil.GetInputPeerID(d.Peer)

		// we can update our access hash state if there is any new peer.
		if err = applyPeers(ctx, manager, d.Entities, id); err != nil {
			log.Warn("failed to apply peer updates", zap.Int64("id", id), zap.Error(err))
		}

		// filter blocked peers
		if _, ok := blocked[id]; ok {
			blockedCount++
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
			nilCount++
			log.Debug("skipping nil dialog",
				zap.Int64("id", id),
				zap.String("peer_type", fmt.Sprintf("%T", d.Peer)))
			continue
		}
		processedCount++

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

	log.Info("dialog processing summary",
		zap.Int("total_fetched", len(allDialogs)),
		zap.Int("blocked", blockedCount),
		zap.Int("nil_dialogs", nilCount),
		zap.Int("processed", processedCount),
		zap.Int("skipped_invalid_peer", skipped),
		zap.Int("non_dialog_types", nonDialogCount),
		zap.Int("final_result", len(result)))

	switch opts.Output {
	case ListOutputTable:
		printTable(log, result)
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

func printTable(log *zap.Logger, result []*Dialog) {
	// Print header
	fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
		"ID",
		"Type",
		"VisibleName",
		"Username",
		"Topics")

	for i, r := range result {
		rowNum := i + 1
		log.Debug("printing table row",
			zap.Int("row", rowNum),
			zap.Int64("id", r.ID),
			zap.String("type", r.Type),
			zap.String("visible_name", r.VisibleName),
			zap.String("username", r.Username),
			zap.Int("topics_count", len(r.Topics)))

		visibleName := r.VisibleName
		if visibleName == "" {
			visibleName = "-"
		}
		username := r.Username
		if username == "" {
			username = "-"
		}

		fmt.Printf("%d\t%s\t%s\t%s\t%s\n",
			r.ID,
			r.Type,
			visibleName,
			username,
			topicsString(r.Topics))
	}

	log.Info("table printing complete", zap.Int("total_rows", len(result)))
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
			logctx.From(ctx).Warn("failed to fetch topics, returning dialog without topics",
				zap.Int64("channel_id", c.ID),
				zap.String("channel_username", c.Username),
				zap.Error(err))
			// Return dialog with empty topics instead of nil
			d.Topics = []Topic{}
		} else {
			d.Topics = topics
		}
	}

	return d
}

// fetchTopics https://github.com/telegramdesktop/tdesktop/blob/4047f1733decd5edf96d125589f128758b68d922/Telegram/SourceFiles/data/data_forum.cpp#L135
func fetchTopics(ctx context.Context, api *tg.Client, c tg.InputChannelClass) ([]Topic, error) {
	res := make([]Topic, 0)
	limit := 100 // why can't we use 500 like tdesktop?
	offsetTopic, offsetID, offsetDate := 0, 0, 0

	for {
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

		for _, tp := range topics.Topics {
			if t, ok := tp.(*tg.ForumTopic); ok {
				res = append(res, Topic{
					ID:    t.ID,
					Title: t.Title,
				})

				offsetTopic = t.ID
			}
		}

		// last page
		if len(topics.Topics) < limit {
			break
		}

		// Update pagination offsets from messages
		// If no messages available, we can't paginate further even if we got full page of topics
		if len(topics.Messages) == 0 {
			break
		}

		if lastMsg, ok := topics.Messages[len(topics.Messages)-1].AsNotEmpty(); ok {
			offsetID, offsetDate = lastMsg.GetID(), lastMsg.GetDate()
		} else {
			// No valid message to use for offset, stop pagination
			break
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
