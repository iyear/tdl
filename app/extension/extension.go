package extension

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/iyear/tdl/pkg/extensions"
)

var (
	colorPrint = func(attrs ...color.Attribute) func(format string, a ...interface{}) {
		return func(format string, a ...interface{}) {
			color.New(attrs...).Print("• ")
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

func Install(ctx context.Context, em *extensions.Manager, target string) error {
	info("installing extension %s...", normalizeExtName(target))

	if err := em.Install(ctx, target); err != nil {
		fail("install extension %s failed: %s", normalizeExtName(target), err)
		return nil
	}

	succ("extension %s installed", normalizeExtName(target))
	return nil
}

func Upgrade(ctx context.Context, em *extensions.Manager, ext string) error {
	exts, err := em.List(ctx, ext == "")
	if err != nil {
		return errors.Wrap(err, "list extensions with metadata")
	}
	if len(exts) == 0 {
		return errors.New("no extensions installed")
	}

	ext = strings.TrimPrefix(ext, extensions.Prefix)

	for _, e := range exts {
		// ext == "": upgrade all extensions
		if ext != "" && e.Name() != ext {
			continue
		}

		info("upgrading %s...", normalizeExtName(e.Name()))

		if err = em.Upgrade(ctx, e); err != nil {
			switch {
			case errors.Is(err, extensions.ErrAlreadyUpToDate):
				succ("extension %s already up-to-date", normalizeExtName(e.Name()))
			case errors.Is(err, extensions.ErrOnlyGitHub):
				fail("extension %s can't be automatically upgraded by tdl", normalizeExtName(e.Name()))
			default:
				fail("upgrade extension %s failed: %s", normalizeExtName(e.Name()), err)
			}

			continue
		}

		if em.DryRun() {
			succ("extension %s will be upgraded", normalizeExtName(e.Name()))
		} else {
			succ("extension %s upgraded", normalizeExtName(e.Name()))
		}
	}

	return nil
}

func Remove(ctx context.Context, em *extensions.Manager, ext string) error {
	exts, err := em.List(ctx, false)
	if err != nil {
		return errors.Wrap(err, "list extensions")
	}

	ext = strings.TrimPrefix(ext, extensions.Prefix)

	for _, e := range exts {
		if ext == e.Name() {
			if err = em.Remove(e); err != nil {
				fail("remove extension %s failed: %s", normalizeExtName(e.Name()), err)
				return nil
			}

			succ("extension %s removed", normalizeExtName(e.Name()))
			return nil
		}
	}

	return fmt.Errorf("no extension matched %q", ext)
}

func normalizeExtName(n string) string {
	if idx := strings.IndexRune(n, '/'); idx >= 0 {
		n = n[idx+1:]
	}
	if !strings.HasPrefix(n, extensions.Prefix) {
		n = extensions.Prefix + n
	}
	return color.New(color.Bold, color.FgCyan).Sprint(n)
}
