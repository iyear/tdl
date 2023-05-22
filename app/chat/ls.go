package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/texpr"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/mattn/go-runewidth"
	"golang.org/x/time/rate"
	"strconv"
	"strings"
	"time"
)

type dialog struct {
	ID          int64   `json:"id" comment:"ID of dialog"`
	Type        string  `json:"type" comment:"Type of dialog. Can be 'user', 'channel' or 'group'"`
	VisibleName string  `json:"visible_name,omitempty" comment:"Title of channel and group, first and last name of user. If empty, output '-'"`
	Username    string  `json:"username,omitempty" comment:"Username of dialog. If empty, output '-'"`
	Topics      []topic `json:"topics,omitempty" comment:"Topics of dialog. If not set, output '-'"`
}

type topic struct {
	ID    int    `json:"id" comment:"ID of topic"`
	Title string `json:"title" comment:"Title of topic"`
}

type Output string

var (
	OutputTable Output = "table"
	OutputJSON  Output = "json"
)

// External designation, different from Telegram mtproto
const (
	DialogGroup   = "group"
	DialogPrivate = "private"
	DialogChannel = "channel"
	DialogUnknown = "unknown"
)

type ListOptions struct {
	Output string
	Filter string
}

func List(ctx context.Context, opts ListOptions) error {
	// align output
	runewidth.EastAsianWidth = false
	runewidth.DefaultCondition.EastAsianWidth = false

	// output available fields
	if opts.Filter == "-" {
		fg := texpr.NewFieldsGetter(nil)
		fields, err := fg.Walk(&dialog{})
		if err != nil {
			return fmt.Errorf("failed to walk fields: %w", err)
		}

		fmt.Print(fg.Sprint(fields, true))
		return nil
	}
	// compile filter
	filter, err := texpr.Compile(opts.Filter)
	if err != nil {
		return fmt.Errorf("failed to compile filter: %w", err)
	}

	c, _, err := tgc.NoLogin(ctx, ratelimit.New(rate.Every(time.Millisecond*400), 2))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		color.Blue("Getting dialogs...")

		dialogs, err := query.GetDialogs(c.API()).BatchSize(100).Collect(ctx)
		if err != nil {
			return err
		}

		blocked, err := utils.Telegram.GetBlockedDialogs(ctx, c.API())
		if err != nil {
			return err
		}

		result := make([]*dialog, 0, len(dialogs))
		for _, d := range dialogs {
			id := utils.Telegram.GetInputPeerID(d.Peer)
			if _, ok := blocked[id]; ok {
				continue
			}

			var r *dialog
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

		switch Output(opts.Output) {
		case OutputTable:
			printTable(result)
		case OutputJSON:
			bytes, err := json.MarshalIndent(result, "", "\t")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}

			fmt.Println(string(bytes))
		default:
			return fmt.Errorf("unknown output: %s", opts.Output)
		}

		return nil
	})
}

func printTable(result []*dialog) {
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

func topicsString(topics []topic) string {
	if len(topics) == 0 {
		return "-"
	}

	s := make([]string, 0, len(topics))
	for _, t := range topics {
		s = append(s, fmt.Sprintf("%d: %s", t.ID, t.Title))
	}

	return strings.Join(s, ", ")
}

func processUser(id int64, entities peer.Entities) *dialog {
	u, ok := entities.User(id)
	if !ok {
		return nil
	}

	return &dialog{
		ID:          u.ID,
		VisibleName: visibleName(u.FirstName, u.LastName),
		Username:    u.Username,
		Type:        DialogPrivate,
		Topics:      nil,
	}
}

func processChannel(ctx context.Context, api *tg.Client, id int64, entities peer.Entities) *dialog {
	c, ok := entities.Channel(id)
	if !ok {
		return nil
	}

	d := &dialog{
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
		req := &tg.ChannelsGetForumTopicsRequest{
			Channel: c.AsInput(),
			Limit:   100,
		}

		topics, err := api.ChannelsGetForumTopics(ctx, req)
		if err != nil {
			return nil
		}

		d.Topics = make([]topic, 0, len(topics.Topics))
		for _, tp := range topics.Topics {
			if t, ok := tp.(*tg.ForumTopic); ok {
				d.Topics = append(d.Topics, topic{
					ID:    t.ID,
					Title: t.Title,
				})
			}
		}
	}

	return d
}

func processChat(id int64, entities peer.Entities) *dialog {
	c, ok := entities.Chat(id)
	if !ok {
		return nil
	}

	return &dialog{
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
