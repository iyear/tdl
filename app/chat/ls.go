package chat

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/query"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/utils"
	"golang.org/x/time/rate"
	"time"
)

func List(ctx context.Context) error {
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

		for _, dialog := range dialogs {
			id := utils.Telegram.GetInputPeerID(dialog.Peer)

			if _, ok := blocked[id]; ok {
				continue
			}

			fmt.Printf("ID: %d, Title: %s, Type: %s\n", id, utils.Telegram.GetPeerName(id, dialog.Entities), utils.Telegram.GetPeerType(id, dialog.Entities))
		}

		return nil
	})
}
