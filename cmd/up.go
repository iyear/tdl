package cmd

import (
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/up"
	"github.com/iyear/tdl/pkg/logger"
)

func NewUpload() *cobra.Command {
	var opts up.Options

	cmd := &cobra.Command{
		Use:     "upload",
		Aliases: []string{"up"},
		Short:   "Upload anything to Telegram",
		RunE: func(cmd *cobra.Command, args []string) error {
			return up.Run(logger.Named(cmd.Context(), "up"), &opts)
		},
	}

	const (
		_chat = "chat"
		path  = "path"
	)
	cmd.Flags().StringVarP(&opts.Chat, _chat, "c", "", "chat id or domain, and empty means 'Saved Messages'")
	cmd.Flags().StringSliceVarP(&opts.Paths, path, "p", []string{}, "dirs or files")
	cmd.Flags().StringSliceVarP(&opts.Excludes, "excludes", "e", []string{}, "exclude the specified file extensions")
	cmd.Flags().BoolVar(&opts.Remove, "rm", false, "remove the uploaded files after uploading")
	cmd.Flags().BoolVar(&opts.Photo, "photo", false, "upload the image as a photo instead of a file")

	// completion and validation
	_ = cmd.MarkFlagRequired(path)

	return cmd
}
