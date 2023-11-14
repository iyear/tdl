package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/fatih/color"

	"github.com/iyear/tdl/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.New().ExecuteContext(ctx); err != nil {
		color.Red("Error: %v", err)
		color.Red("%+v", err)
		os.Exit(1)
	}
}
