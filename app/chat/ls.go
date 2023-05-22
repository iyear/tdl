package chat

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/mattn/go-runewidth"
	"golang.org/x/time/rate"
	"strconv"
	"strings"
	"time"
)

type dialog struct {
	ID          int64   `json:"id"`
	VisibleName string  `json:"visible_name"`
	Username    string  `json:"username"`
	Type        string  `json:"type"`
	Topics      []topic `json:"topics,omitempty"`
}

type topic struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func List(ctx context.Context) error {
	// align output
	runewidth.EastAsianWidth = false
	runewidth.DefaultCondition.EastAsianWidth = false

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

		fmt.Printf("%s %s %s %s %s\n",
			trunc("ID", 10),
			trunc("Type", 8),
			trunc("VisibleName", 20),
			trunc("Username", 20),
			"Topics")

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

			if r == nil {
				continue
			}

			fmt.Printf("%s %s %s %s %s\n",
				trunc(strconv.FormatInt(r.ID, 10), 10),
				trunc(r.Type, 8),
				trunc(r.VisibleName, 20),
				trunc(r.Username, 20),
				topicsString(r.Topics))
		}

		return nil
	})
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
		VisibleName: u.FirstName + " " + u.LastName,
		Username:    u.Username,
		Type:        consts.DialogPrivate,
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
		d.Type = consts.DialogChannel
	case c.Megagroup, c.Gigagroup:
		d.Type = consts.DialogGroup
	default:
		d.Type = consts.DialogUnknown
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
		Type:        consts.DialogGroup,
		Topics:      nil,
	}
}
