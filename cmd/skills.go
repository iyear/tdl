package cmd

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/docs"
	"github.com/iyear/tdl/pkg/consts"
)

func NewSkills() *cobra.Command {
	var (
		dir  string
		name string
	)

	cmd := &cobra.Command{
		Use:     "skills",
		Aliases: []string{"skill"},
		Short:   "Install agent skills for tdl",
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir := filepath.Join(dir, name)

			if err := multierr.Combine(
				os.RemoveAll(skillDir), // clean up existing skill dir
				os.CopyFS(skillDir, docs.SkillMd),
				os.CopyFS(filepath.Join(skillDir, "references"), docs.Skills),
			); err != nil {
				return errors.Wrap(err, "create skills")
			}

			color.Green("Skill installed to %q", skillDir)

			return nil
		},
	}

	cmd.Flags().StringVarP(&dir, "directory", "d", filepath.Join(consts.HomeDir, ".agents", "skills"), "directory path for skills")
	cmd.Flags().StringVar(&name, "name", "tdl", "skill name")

	return cmd
}
