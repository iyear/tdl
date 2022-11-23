package login

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
)

const tdata = "tdata"

func Desktop(ctx context.Context, desktop, passcode string) error {
	ns := viper.GetString(consts.FlagNamespace)

	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})

	// process path that points to Telegram executable file
	stat, err := os.Stat(desktop)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		desktop = filepath.Dir(desktop)
	}

	color.Blue("Importing session from desktop client: %s", desktop)

	if filepath.Base(desktop) != tdata {
		desktop = filepath.Join(desktop, tdata)
	}
	accounts, err := tdesktop.Read(desktop, []byte(passcode))
	if err != nil {
		return err
	}

	infos := make([]string, 0, len(accounts))
	infoMap := make(map[string]tdesktop.Account)
	for _, acc := range accounts {
		id := strconv.FormatUint(acc.Authorization.UserID, 10)
		infos = append(infos, id)
		infoMap[id] = acc
	}

	fmt.Println()
	acc := ""
	prompt := &survey.Select{
		Message: "Choose a user id:",
		Options: infos,
		Help:    "You can get user id from @userinfobot",
	}
	if err = survey.AskOne(prompt, &acc); err != nil {
		return err
	}

	data, err := session.TDesktopSession(infoMap[acc])
	if err != nil {
		return err
	}

	loader := &session.Loader{Storage: storage.NewSession(kvd, true)}
	if err = loader.Save(ctx, data); err != nil {
		return err
	}

	if err = kvd.Set(key.App(), []byte(consts.AppDesktop)); err != nil {
		return err
	}

	color.Green("Import %s successfully to '%s' namespace!", acc, ns)
	return nil
}
