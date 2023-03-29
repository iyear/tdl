package main

import (
	"context"
	"github.com/fatih/color"
	"github.com/iyear/tdl/cmd"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.New().ExecuteContext(ctx); err != nil {
		color.Red("%v", err)
	}
}
