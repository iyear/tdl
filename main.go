package main

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		color.Red("%v", err)
	}
}
