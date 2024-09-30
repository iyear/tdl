package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/migrate"
	"github.com/iyear/tdl/pkg/kv"
)

func NewBackup() *cobra.Command {
	var dst string

	cmd := &cobra.Command{
		Use:     "backup",
		Short:   "Backup your data",
		GroupID: groupAccount.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dst == "" {
				dst = fmt.Sprintf("%s.backup.tdl", time.Now().Format("2006-01-02-15_04_05"))
			}

			return migrate.Backup(cmd.Context(), dst)
		},
	}

	cmd.Flags().StringVarP(&dst, "dst", "d", "", "destination file path. Default: <date>.backup.tdl")

	return cmd
}

func NewRecover() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:     "recover",
		Short:   "Recover your data",
		GroupID: groupAccount.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrate.Recover(cmd.Context(), file)
		},
	}

	const fileFlag = "file"

	cmd.Flags().StringVarP(&file, fileFlag, "f", "", "backup file path")

	// completion and validation
	_ = cmd.RegisterFlagCompletionFunc(fileFlag, completeExtFiles("tdl"))
	_ = cmd.MarkFlagRequired(fileFlag)

	return cmd
}

func NewMigrate() *cobra.Command {
	var to map[string]string

	cmd := &cobra.Command{
		Use:     "migrate",
		Short:   "Migrate your current data to another storage",
		GroupID: groupAccount.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrate.Migrate(cmd.Context(), to)
		},
	}

	cmd.Flags().StringToStringVar(&to, "to", map[string]string{},
		fmt.Sprintf("destination storage options, format: type=driver,key1=value1,key2=value2. Available drivers: [%s]",
			strings.Join(kv.DriverNames(), ",")))

	return cmd
}
