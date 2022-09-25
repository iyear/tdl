package dl

import (
	"context"
	"fmt"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/spf13/viper"
)

func Run(ctx context.Context, urls, files []string) error {
	c, _, err := tgc.NoLogin()
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

		umsgs, err := parseURLs(ctx, c.API(), urls)
		if err != nil {
			return err
		}

		fmsgs, err := parseFiles(ctx, c.API(), files)
		if err != nil {
			return err
		}

		return downloader.New(c.API(), viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), newIter(c.API(), umsgs, fmsgs)).
			Download(ctx, viper.GetInt(consts.FlagLimit))
	})
}
