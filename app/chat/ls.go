package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
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

	// Try to fetch dialogs with normal batch size first
	// If it fails due to a problematic dialog (deleted/inaccessible channel),
	// we'll retry with smaller batches to isolate and skip the problem
	elems, err := fetchDialogsWithFallback(ctx, c.API(), log)
	if err != nil {
		return fmt.Errorf("failed to fetch dialogs: %w", err)
	}

	blocked, err := tutil.GetBlockedDialogs(ctx, c.API())
	if err != nil {
		return err
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
	result := make([]*Dialog, 0, len(elems))
	for _, d := range elems {
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

		if len(topics.Messages) > 0 {
			if lastMsg, ok := topics.Messages[len(topics.Messages)-1].AsNotEmpty(); ok {
				offsetID, offsetDate = lastMsg.GetID(), lastMsg.GetDate()
			}
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

// fetchDialogsWithFallback attempts to fetch dialogs with smart fallback:
// 1. Try with optimal batch size (100) - fast for normal cases
// 2. If that fails, retry with smaller batches (10) to isolate problematic dialogs
// 3. Continue with partial results if some dialogs are inaccessible
func fetchDialogsWithFallback(ctx context.Context, api *tg.Client, log *zap.Logger) ([]dialogs.Elem, error) {
	// First attempt: use large batch size for best performance
	elems, err := fetchDialogsWithBatchSize(ctx, api, 100)
	if err == nil {
		return elems, nil
	}

	// If we got nothing with large batches, try smaller batches to work around problematic dialogs
	log.Warn("failed to fetch all dialogs with optimal batch size, retrying with smaller batches",
		zap.Error(err))
	
	elems, err = fetchDialogsWithBatchSize(ctx, api, 10)
	if err != nil {
		if len(elems) > 0 {
			// Got partial results before hitting error
			log.Warn("fetched partial dialogs due to inaccessible/deleted channels",
				zap.Int("fetched", len(elems)),
				zap.Error(err))
			return elems, nil
		}
		// Complete failure
		return nil, err
	}

	return elems, nil
}

// fetchDialogsWithBatchSize fetches dialogs with specified batch size
func fetchDialogsWithBatchSize(ctx context.Context, api *tg.Client, batchSize int) ([]dialogs.Elem, error) {
	iter := query.GetDialogs(api).BatchSize(batchSize).Iter()
	elems := []dialogs.Elem{}
	
	for iter.Next(ctx) {
		elems = append(elems, iter.Value())
	}
	
	return elems, iter.Err()
}
