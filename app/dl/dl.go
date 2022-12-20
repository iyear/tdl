package dl

import (
	"context"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/spf13/viper"
)

func Run(ctx context.Context, dir string, rewriteExt bool, template string, urls, files, include, exclude []string) error {
	c, kvd, err := tgc.NoLogin()
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		umsgs, err := parseURLs(ctx, c.API(), kvd, urls)
		if err != nil {
			return err
		}

		fmsgs, err := parseFiles(ctx, c.API(), kvd, files)
		if err != nil {
			return err
		}

		it, err := newIter(c.API(), kvd, template, include, exclude, umsgs, fmsgs)
		if err != nil {
			return err
		}
		return downloader.New(c.API(), dir, rewriteExt, viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), it).
			Download(ctx, viper.GetInt(consts.FlagLimit))
	})
}
