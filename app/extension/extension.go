package extension

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/iyear/tdl/pkg/extensions"
)

var (
	colorPrint = func(attrs ...color.Attribute) func(padding int, format string, a ...interface{}) {
		return func(padding int, format string, a ...interface{}) {
			color.New(attrs...).Print(strings.Repeat("  ", padding) + "• ")
			fmt.Printf(format+"\n", a...)
		}
	}
	info = colorPrint(color.FgBlue, color.Bold)
	succ = colorPrint(color.FgGreen, color.Bold)
	fail = colorPrint(color.FgRed, color.Bold)
)

func List(ctx context.Context, em *extensions.Manager) error {
	exts, err := em.List(ctx, false)
	if err != nil {
		return errors.New("list extensions failed")
	}

	tb := table.NewWriter()

	style := table.StyleColoredDark
	tb.SetStyle(style)

	tb.AppendHeader(table.Row{"NAME", "AUTHOR", "VERSION"})
	for _, e := range exts {
		tb.AppendRow(table.Row{normalizeExtName(e.Name()), e.Owner(), e.CurrentVersion()})
	}

	fmt.Println(tb.Render())

	return nil
}

func Install(ctx context.Context, em *extensions.Manager, targets []string, force bool) error {
	for _, target := range targets {
		info(0, "installing extension %s...", normalizeExtName(target))

		if err := em.Install(ctx, target, force); err != nil {
			fail(1, "install extension %s failed: %s", normalizeExtName(target), err)
			continue
		}

		if em.DryRun() {
			succ(1, "extension %s will be installed", normalizeExtName(target))
		} else {
			succ(1, "extension %s installed", normalizeExtName(target))
		}
	}

	return nil
}

func Upgrade(ctx context.Context, em *extensions.Manager, targets []string) error {
	upgradeAll := len(targets) == 0

	exts, err := em.List(ctx, upgradeAll)
	if err != nil {
		return errors.Wrap(err, "list extensions with metadata")
	}
	if len(exts) == 0 {
		return errors.New("no extensions installed")
	}

	extMap := make(map[string]extensions.Extension)
	for _, e := range exts {
		extMap[e.Name()] = e
		if upgradeAll {
			targets = append(targets, e.Name())
		}
	}

	for _, target := range targets {
		e, ok := extMap[strings.TrimPrefix(target, extensions.Prefix)]
		if !ok {
			fail(0, "extension %s not found", normalizeExtName(target))
			continue
		}

		info(0, "upgrading %s...", normalizeExtName(e.Name()))

		if err = em.Upgrade(ctx, e); err != nil {
			switch {
			case errors.Is(err, extensions.ErrAlreadyUpToDate):
				succ(1, "extension %s already up-to-date", normalizeExtName(e.Name()))
			case errors.Is(err, extensions.ErrOnlyGitHub):
				fail(1, "extension %s can't be automatically upgraded by tdl", normalizeExtName(e.Name()))
			default:
				fail(1, "upgrade extension %s failed: %s", normalizeExtName(e.Name()), err)
			}

			continue
		}

		if em.DryRun() {
			succ(1, "extension %s will be upgraded", normalizeExtName(e.Name()))
		} else {
			succ(1, "extension %s upgraded", normalizeExtName(e.Name()))
		}
	}

	return nil
}

func Remove(ctx context.Context, em *extensions.Manager, targets []string) error {
	exts, err := em.List(ctx, false)
	if err != nil {
		return errors.Wrap(err, "list extensions")
	}

	extMap := make(map[string]extensions.Extension)
	for _, e := range exts {
		extMap[e.Name()] = e
	}

	for _, target := range targets {
		e, ok := extMap[strings.TrimPrefix(target, extensions.Prefix)]
		if !ok {
			fail(0, "extension %s not found", normalizeExtName(target))
			continue
		}

		if err = em.Remove(e); err != nil {
			fail(0, "remove extension %s failed: %s", normalizeExtName(e.Name()), err)
			continue
		}

		if em.DryRun() {
			succ(0, "extension %s will be removed", normalizeExtName(e.Name()))
		} else {
			succ(0, "extension %s removed", normalizeExtName(e.Name()))
		}
	}

	return nil
}

func normalizeExtName(n string) string {
	if idx := strings.IndexRune(n, '/'); idx >= 0 {
		n = n[idx+1:]
	}
	if !strings.HasPrefix(n, extensions.Prefix) {
		n = extensions.Prefix + n
	}
	n = strings.TrimSuffix(n, filepath.Ext(n))
	return color.New(color.Bold, color.FgCyan).Sprint(n)
}
