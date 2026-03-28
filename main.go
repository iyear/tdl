package main

import (
	"context"
	"os"
	"os/signal"

	surveyterm "github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	bbolterrors "go.etcd.io/bbolt/errors"

	"github.com/iyear/tdl/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	humanizeErrors := map[error]string{
		bbolterrors.ErrTimeout:  "Current database is used by another process, please terminate it first",
		surveyterm.InterruptErr: "Interrupted",
	}

	if err := cmd.New().ExecuteContext(ctx); err != nil {
		for e, m := range humanizeErrors {
			if errors.Is(err, e) {
				color.Red("%s", m)
				os.Exit(1)
			}
		}

		color.Red("Error: %+v", err)
		os.Exit(1)
	}
}
