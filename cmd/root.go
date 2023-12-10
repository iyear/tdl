package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/tclient"
)

func New() *cobra.Command {
	driverTypeKey := "type"

	cmd := &cobra.Command{
		Use:           "tdl",
		Short:         "Telegram Downloader, but more than a downloader",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// init logger
			debug, level := viper.GetBool(consts.FlagDebug), zap.InfoLevel
			if debug {
				level = zap.DebugLevel
			}
			cmd.SetContext(logger.With(cmd.Context(),
				logger.New(level, filepath.Join(consts.LogPath, "latest.log"))))

			ns := viper.GetString(consts.FlagNamespace)
			if ns != "" {
				logger.From(cmd.Context()).Info("Namespace",
					zap.String("namespace", ns))
			}

			// check storage flag
			storageOpts := viper.GetStringMapString(consts.FlagStorage)
			driver, err := kv.ParseDriver(storageOpts[driverTypeKey])
			if err != nil {
				return errors.Wrap(err, "parse driver")
			}
			delete(storageOpts, driverTypeKey)

			opts := make(map[string]any)
			for k, v := range storageOpts {
				opts[k] = v
			}
			storage, err := kv.New(driver, opts)
			if err != nil {
				return errors.Wrap(err, "create kv storage")
			}

			cmd.SetContext(kv.With(cmd.Context(), storage))
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return multierr.Combine(
				kv.From(cmd.Context()).Close(),
				logger.From(cmd.Context()).Sync(),
			)
		},
	}

	cmd.AddCommand(NewVersion(), NewLogin(), NewDownload(), NewForward(),
		NewChat(), NewUpload(), NewBackup(), NewRecover(), NewGen())

	cmd.PersistentFlags().StringToString(consts.FlagStorage, map[string]string{
		driverTypeKey: kv.DriverLegacy.String(),
		"path":        consts.KVPath,
	}, fmt.Sprintf("storage options, format: type=driver,key1=value1,key2=value2. Available drivers: [%s]",
		strings.Join(kv.DriverNames(), ",")))

	cmd.PersistentFlags().String(consts.FlagProxy, "", "proxy address, format: protocol://username:password@host:port")
	cmd.PersistentFlags().StringP(consts.FlagNamespace, "n", "", "namespace for Telegram session")
	cmd.PersistentFlags().Bool(consts.FlagDebug, false, "enable debug mode")

	cmd.PersistentFlags().IntP(consts.FlagPartSize, "s", 512*1024, "part size for transfer, max is 512*1024")
	cmd.PersistentFlags().IntP(consts.FlagThreads, "t", 4, "max threads for transfer one item")
	cmd.PersistentFlags().IntP(consts.FlagLimit, "l", 2, "max number of concurrent tasks")
	cmd.PersistentFlags().Int(consts.FlagPoolSize, 8, "specify the size of the DC pool, zero means infinity")

	cmd.PersistentFlags().String(consts.FlagNTP, "", "ntp server host, if not set, use system time")
	cmd.PersistentFlags().Duration(consts.FlagReconnectTimeout, 2*time.Minute, "Telegram client reconnection backoff timeout, infinite if set to 0") // #158

	cmd.PersistentFlags().String(consts.FlagTest, "", "use test Telegram client, only for developer")

	// completion
	_ = cmd.RegisterFlagCompletionFunc(consts.FlagNamespace, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		engine := kv.From(cmd.Context())
		ns, err := engine.Namespaces()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return ns, cobra.ShellCompDirectiveNoFileComp
	})

	_ = viper.BindPFlags(cmd.PersistentFlags())

	viper.SetEnvPrefix("tdl")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	return cmd
}

type completeFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

func completeExtFiles(ext ...string) completeFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		files := make([]string, 0)
		for _, e := range ext {
			f, err := filepath.Glob(toComplete + "*." + e)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}
			files = append(files, f...)
		}

		return files, cobra.ShellCompDirectiveFilterDirs
	}
}

func tRun(ctx context.Context, login bool, f func(ctx context.Context, c *telegram.Client, kvd kv.KV) error, middlewares ...telegram.Middleware) error {
	// init tclient kv
	kvd, err := kv.From(ctx).Open(viper.GetString(consts.FlagNamespace))
	if err != nil {
		return errors.Wrap(err, "open kv storage")
	}
	o := tclient.Options{
		KV:               kvd,
		Proxy:            viper.GetString(consts.FlagProxy),
		NTP:              viper.GetString(consts.FlagNTP),
		ReconnectTimeout: viper.GetDuration(consts.FlagReconnectTimeout),
		Test:             viper.GetString(consts.FlagTest) != "",
	}

	client, err := tclient.New(ctx, o, login, middlewares...)
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	return tclient.Run(ctx, client, func(ctx context.Context) error {
		return f(ctx, client, kvd)
	})
}
