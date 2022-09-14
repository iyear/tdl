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
)

func Run(ctx context.Context, ns, proxy string, partSize, threads, limit int, paths, excludes []string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})
	if err != nil {
		return err
	}

	files, err := walk(paths, excludes)
	if err != nil {
		return err
	}

	color.Blue("Files count: %d", len(files))

	c := tgc.New(proxy, kvd, false, floodwait.NewSimpleWaiter())
	return c.Run(ctx, func(ctx context.Context) error {
		status, err := c.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		return uploader.New(c.API(), partSize, threads, newIter(files)).Upload(ctx, limit)
	})
}
