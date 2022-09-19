package login

import (
	"context"
	"fmt"
	"github.com/dpastoor/go-input"
	"github.com/fatih/color"
	"github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/spf13/viper"
	"path/filepath"
	"strconv"
)

const tdata = "tdata"

func Desktop(ctx context.Context, desktop string) error {
	ns := viper.GetString(consts.FlagNamespace)

	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})

	color.Blue("Importing session from desktop client: %s", desktop)

	if filepath.Base(desktop) != tdata {
		desktop = filepath.Join(desktop, tdata)
	}
	accounts, err := tdesktop.Read(desktop, nil)
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

	color.Blue("You can get user id from @userinfobot")
	fmt.Println()

	acc, err := input.DefaultUI().Select(color.BlueString("Select a user id:"), infos, &input.Options{
		Loop:     true,
		Required: true,
	})
	if err != nil {
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

	color.Green("Import %s successfully to '%s' namespace!", acc)
	return nil
}
