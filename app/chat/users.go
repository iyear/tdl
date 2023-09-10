package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/go-faster/jx"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/tg"
	"github.com/jedib0t/go-pretty/v6/progress"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"

	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

type UsersOptions struct {
	Chat   string
	Output string
	Raw    bool
}

type User struct {
	ID        int64  `json:"id"`
	Bot       bool   `json:"bot"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func Users(ctx context.Context, opts *UsersOptions) error {
	c, kvd, err := tgc.NoLogin(ctx, ratelimit.New(rate.Every(rateInterval), rateBucket))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		var peer peers.Peer

		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
		if opts.Chat == "" { // defaults to me(saved messages)
			return fmt.Errorf("missing domain id")
		}

		peer, err = utils.Telegram.GetInputPeer(ctx, manager, opts.Chat)
		if err != nil {
			return fmt.Errorf("failed to get peer: %w", err)
		}

		color.Cyan("Occasional suspensions are due to Telegram rate limitations, please wait a moment.")
		fmt.Println()

		pw := prog.New(progress.FormatNumber)
		pw.SetUpdateFrequency(200 * time.Millisecond)
		pw.Style().Visibility.TrackerOverall = false
		pw.Style().Visibility.ETA = false
		pw.Style().Visibility.Percentage = false

		tracker := prog.AppendTracker(pw, progress.FormatNumber, fmt.Sprintf("%s-%d", peer.VisibleName(), peer.ID()), 0)

		go pw.Render()

		ch, ok := peer.(peers.Channel)
		if !ok {
			return fmt.Errorf("invalid type of chat. channels are supported only")
		}
		usersList := []*tg.User{}

		iter := participants.NewIterator(query.NewQuery(c.API()).GetParticipants(ch.InputChannel()), 100)
		for iter.Next(ctx) {
			el := iter.Value()
			us, ok := el.User()
			if !ok {
				continue
			}

			usersList = append(usersList, us)
		}

		if err = iter.Err(); err != nil {
			return err
		}

		f, err := os.Create(opts.Output)
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&rerr, multierr.Close(f))

		enc := jx.NewStreamingEncoder(f, 512)
		defer multierr.AppendInvoke(&rerr, multierr.Close(enc))

		enc.ObjStart()
		defer enc.ObjEnd()
		enc.Field("id", func(e *jx.Encoder) { e.Int64(peer.ID()) })

		enc.FieldStart("users")
		var output any = usersList
		if !opts.Raw {
			users := make([]User, len(usersList))
			for i := 0; i < len(usersList); i++ {
				convertTelegramUser(&users[i], usersList[i])
			}

			output = users
		}

		buf, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		enc.Raw(buf)

		tracker.MarkAsDone()
		prog.Wait(pw)
		return nil
	})
}

func convertTelegramUser(dstUser *User, tgUser *tg.User) {
	dstUser.ID = tgUser.ID
	dstUser.Bot = tgUser.Bot
	dstUser.FirstName = tgUser.FirstName
	dstUser.LastName = tgUser.LastName
	dstUser.Phone = tgUser.Phone
	dstUser.Username = tgUser.Username
}
