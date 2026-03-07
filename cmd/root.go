package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	tclientcore "github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/core/util/logutil"
	"github.com/iyear/tdl/core/util/netutil"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/extensions"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
)

var (
	defaultBoltPath = filepath.Join(consts.DataDir, "data")

	DefaultLegacyStorage = map[string]string{
		kv.DriverTypeKey: kv.DriverLegacy.String(),
		"path":           filepath.Join(consts.DataDir, "data.kv"),
	}
	DefaultBoltStorage = map[string]string{
		kv.DriverTypeKey: kv.DriverBolt.String(),
		"path":           defaultBoltPath,
	}
)

// command groups
var (
	groupAccount = &cobra.Group{
		ID:    "account",
		Title: "Account related",
	}
	groupTools = &cobra.Group{
		ID:    "tools",
		Title: "Tools",
	}
	groupExtensions = &cobra.Group{
		ID:    "extensions",
		Title: "Extensions",
	}
)

func New() *cobra.Command {
	// allow PersistentPreRun to be called for every command
	cobra.EnableTraverseRunHooks = true

	em := extensions.NewManager(consts.ExtensionsPath)

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
			cmd.SetContext(logctx.With(cmd.Context(),
				logutil.New(level, filepath.Join(consts.LogPath, "latest.log"))))

			ns := viper.GetString(consts.FlagNamespace)
			if ns != "" {
				logctx.From(cmd.Context()).Info("Namespace",
					zap.String("namespace", ns))
			}

			// v0.14.0: default storage changed from legacy to bolt, so we need to auto migrate to keep compatibility
			if !cmd.Flags().Lookup(consts.FlagStorage).Changed && !fsutil.PathExists(defaultBoltPath) {
				if err := migrateLegacyToBolt(); err != nil {
					return errors.Wrap(err, "migrate legacy to bolt")
				}
			}

			stg, err := kv.NewWithMap(viper.GetStringMapString(consts.FlagStorage))
			if err != nil {
				return errors.Wrap(err, "create kv storage")
			}

			cmd.SetContext(kv.With(cmd.Context(), stg))

			// extension manager client proxy
			var dialer proxy.ContextDialer = proxy.Direct
			if p := viper.GetString(consts.FlagProxy); p != "" {
				if t, err := netutil.NewProxy(p); err == nil {
					dialer = t
				}
			}
			em.SetClient(&http.Client{Transport: &http.Transport{
				DialContext: dialer.DialContext,
			}})

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return multierr.Combine(
				kv.From(cmd.Context()).Close(),
				logctx.From(cmd.Context()).Sync(),
			)
		},
	}

	coloredcobra.Init(&coloredcobra.Config{
		RootCmd:         cmd,
		Headings:        coloredcobra.HiCyan + coloredcobra.Bold + coloredcobra.Underline,
		Commands:        coloredcobra.HiGreen + coloredcobra.Bold,
		CmdShortDescr:   coloredcobra.None,
		ExecName:        coloredcobra.Bold,
		Flags:           coloredcobra.Bold + coloredcobra.Yellow,
		FlagsDataType:   coloredcobra.Blue,
		FlagsDescr:      coloredcobra.None,
		Aliases:         coloredcobra.None,
		Example:         coloredcobra.None,
		NoExtraNewlines: true,
		NoBottomNewline: true,
	})

	cmd.AddGroup(groupAccount, groupTools, groupExtensions)

	cmd.AddCommand(NewVersion(), NewLogin(), NewDownload(), NewForward(),
		NewChat(), NewUpload(), NewBackup(), NewRecover(), NewMigrate(),
		NewGen(), NewExtension(em))

	// append extension command to root
	exts, _ := em.List(context.Background(), false)
	for _, e := range exts {
		cmd.AddCommand(NewExtensionCmd(em, e, os.Stdin, os.Stdout, os.Stderr))
	}

	cmd.PersistentFlags().StringToString(consts.FlagStorage,
		DefaultBoltStorage,
		fmt.Sprintf("storage options, format: type=driver,key1=value1,key2=value2. Available drivers: [%s]",
			strings.Join(kv.DriverNames(), ",")))

	cmd.PersistentFlags().String(consts.FlagProxy, "", "proxy address, format: protocol://username:password@host:port")
	cmd.PersistentFlags().StringP(consts.FlagNamespace, "n", "default", "namespace for Telegram session")
	cmd.PersistentFlags().Bool(consts.FlagDebug, false, "enable debug mode")

	cmd.PersistentFlags().IntP(consts.FlagPartSize, "s", 512*1024, "part size for transfer")
	_ = cmd.PersistentFlags().MarkDeprecated(consts.FlagPartSize, "part size has been set to maximum by default, this flag will be removed in the future")

	cmd.PersistentFlags().IntP(consts.FlagThreads, "t", 4, "max threads for transfer one item")
	cmd.PersistentFlags().IntP(consts.FlagLimit, "l", 2, "max number of concurrent tasks")
	cmd.PersistentFlags().Int(consts.FlagPoolSize, 8, "specify the size of the DC pool, zero means infinity")
	cmd.PersistentFlags().Duration(consts.FlagDelay, 0, "delay between each task, zero means no delay")
	cmd.PersistentFlags().Bool(consts.FlagProgressPS, true, "show pinned CPU/memory/goroutines progress info")

	cmd.PersistentFlags().String(consts.FlagNTP, "", "ntp server host, if not set, use system time")
	cmd.PersistentFlags().Duration(consts.FlagReconnectTimeout, 5*time.Minute, "Telegram client reconnection backoff timeout, infinite if set to 0") // #158

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

	// extension command format: <global-flags> <extension-name> <extension-flags>,
	// which means parse args layer by layer. But common command flags are flat.
	// To keep compatibility, we only set TraverseChildren to true for extension
	// command instead of other commands.
	foundCmd, _, err := cmd.Find(os.Args[1:])
	if err == nil && foundCmd.GroupID == groupExtensions.ID {
		cmd.TraverseChildren = true // allow global config to be parsed before extension command is executed
	}

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

func tOptions(ctx context.Context) (tclient.Options, error) {
	// init tclient kv
	kvd, err := kv.From(ctx).Open(viper.GetString(consts.FlagNamespace))
	if err != nil {
		return tclient.Options{}, errors.Wrap(err, "open kv storage")
	}
	o := tclient.Options{
		KV:               kvd,
		Proxy:            viper.GetString(consts.FlagProxy),
		NTP:              viper.GetString(consts.FlagNTP),
		ReconnectTimeout: viper.GetDuration(consts.FlagReconnectTimeout),
		UpdateHandler:    nil,
	}

	return o, nil
}

func tRun(ctx context.Context, f func(ctx context.Context, c *telegram.Client, kvd storage.Storage) error, middlewares ...telegram.Middleware) error {
	o, err := tOptions(ctx)
	if err != nil {
		return errors.Wrap(err, "build telegram options")
	}

	client, err := tclient.New(ctx, o, false, middlewares...)
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	return tclientcore.RunWithAuth(ctx, client, func(ctx context.Context) error {
		return f(ctx, client, o.KV)
	})
}

func migrateLegacyToBolt() (rerr error) {
	legacy, err := kv.NewWithMap(DefaultLegacyStorage)
	if err != nil {
		return errors.Wrap(err, "create legacy kv storage")
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(legacy))

	bolt, err := kv.NewWithMap(DefaultBoltStorage)
	if err != nil {
		return errors.Wrap(err, "create bolt kv storage")
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(bolt))

	meta, err := legacy.MigrateTo()
	if err != nil {
		return errors.Wrap(err, "migrate legacy to bolt")
	}

	return bolt.MigrateFrom(meta)
}
