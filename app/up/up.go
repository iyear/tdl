package up

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/spf13/viper"
)

func Run(ctx context.Context, paths, excludes []string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   viper.GetString(consts.FlagNamespace),
	})
	if err != nil {
		return err
	}

	files, err := walk(paths, excludes)
	if err != nil {
		return err
	}

	color.Blue("Files count: %d", len(files))

	c, err := tgc.New(viper.GetString(consts.FlagProxy), kvd, false, floodwait.NewSimpleWaiter())
	if err != nil {
		return err
	}

	return c.Run(ctx, func(ctx context.Context) error {
		status, err := c.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		return uploader.New(c.API(), viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), newIter(files)).
			Upload(ctx, viper.GetInt(consts.FlagLimit))
	})
}
