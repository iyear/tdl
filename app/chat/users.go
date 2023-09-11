package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram/peers"
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
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func Users(ctx context.Context, opts UsersOptions) error {
	c, kvd, err := tgc.NoLogin(ctx, ratelimit.New(rate.Every(rateInterval), rateBucket))
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(c.API())
		if opts.Chat == "" {
			return fmt.Errorf("missing domain id")
		}

		peer, err := utils.Telegram.GetInputPeer(ctx, manager, opts.Chat)
		if err != nil {
			return fmt.Errorf("failed to get peer: %w", err)
		}

		ch, ok := peer.(peers.Channel)
		if !ok {
			return fmt.Errorf("invalid type of chat. channels/groups are supported only")
		}

		color.Cyan("Occasional suspensions are due to Telegram rate limitations, please wait a moment.")
		fmt.Println()

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

		pw := prog.New(progress.FormatNumber)
		pw.SetUpdateFrequency(200 * time.Millisecond)
		pw.Style().Visibility.TrackerOverall = false
		pw.Style().Visibility.ETA = true
		pw.Style().Visibility.Percentage = true

		go pw.Render()

		builder := func() *participants.GetParticipantsQueryBuilder {
			return participants.NewQueryBuilder(c.API()).
				GetParticipants(ch.InputChannel()).
				BatchSize(100)
		}

		fields := map[string]*participants.GetParticipantsQueryBuilder{
			"users":  builder(),
			"admins": builder().Admins(),
			"kicked": builder().Kicked(""),
			"banned": builder().Banned(""),
			"bots":   builder().Bots(),
		}

		for field, query := range fields {
			iter := query.Iter()
			if err = outputUsers(ctx, pw, peer, enc, field, iter, opts.Raw); err != nil {
				return fmt.Errorf("failed to output %s: %w", field, err)
			}
		}

		prog.Wait(pw)
		return nil
	})
}

func outputUsers(ctx context.Context,
	pw progress.Writer,
	peer peers.Peer,
	enc *jx.Encoder,
	field string,
	iter *participants.Iterator,
	raw bool,
) error {
	total, err := iter.Total(ctx)
	if err != nil {
		return errors.Wrap(err, "get total count")
	}

	tracker := prog.AppendTracker(pw,
		progress.FormatNumber,
		fmt.Sprintf("%s-%d-%s", peer.VisibleName(), peer.ID(), field),
		int64(total))

	enc.FieldStart(field)
	enc.ArrStart()
	defer enc.ArrEnd()

	for iter.Next(ctx) {
		el := iter.Value()
		u, ok := el.User()
		if !ok {
			continue
		}

		var output any = u
		if !raw {
			output = convertTelegramUser(u)
		}

		buf, err := json.Marshal(output)
		if err != nil {
			return errors.Wrap(err, "marshal user")
		}

		enc.Raw(buf)

		tracker.Increment(1)
	}

	if err = iter.Err(); err != nil {
		return err
	}

	tracker.MarkAsDone()
	return nil
}

func convertTelegramUser(u *tg.User) User {
	var dst User
	dst.ID = u.ID
	dst.Username = u.Username
	dst.Phone = u.Phone
	dst.FirstName = u.FirstName
	dst.LastName = u.LastName
	return dst
}
