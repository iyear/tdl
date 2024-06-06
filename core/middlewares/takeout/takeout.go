package takeout

import (
	"context"

	"github.com/gotd/td/tg"
)

func Takeout(ctx context.Context, invoker tg.Invoker) (int64, error) {
	req := &tg.AccountInitTakeoutSessionRequest{
		Contacts:          true,
		MessageUsers:      true,
		MessageChats:      true,
		MessageMegagroups: true,
		MessageChannels:   true,
		Files:             true,
		FileMaxSize:       4000 * 1024 * 1024,
	}
	req.SetFlags()

	session, err := tg.NewClient(invoker).AccountInitTakeoutSession(ctx, req)
	if err != nil {
		return 0, err
	}
	return session.ID, nil
}

// UnTakeout should be called with takeout wrapper invoker
func UnTakeout(ctx context.Context, invoker tg.Invoker) error {
	req := &tg.AccountFinishTakeoutSessionRequest{Success: true}
	req.SetFlags()

	_, err := tg.NewClient(invoker).AccountFinishTakeoutSession(ctx, req)
	return err
}
