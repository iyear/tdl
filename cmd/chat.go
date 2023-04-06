package cmd

import (
	"fmt"
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/spf13/cobra"
	"math"
	"strings"
)

func NewChat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "A set of chat tools",
	}

	cmd.AddCommand(NewChatList(), NewChatExport())

	return cmd
}

func NewChatList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List your chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return chat.List(logger.Named(cmd.Context(), "ls"))
		},
	}

	return cmd
}

func NewChatExport() *cobra.Command {
	var opts chat.ExportOptions

	cmd := &cobra.Command{
		Use:   "export",
		Short: "export messages from (protected) chat for download",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch opts.Type {
			case chat.ExportTypeTime, chat.ExportTypeID:
				// set default value
				switch len(opts.Input) {
				case 0:
					opts.Input = []int{0, math.MaxInt}
				case 1:
					opts.Input = append(opts.Input, math.MaxInt)
				}

				if len(opts.Input) != 2 {
					return fmt.Errorf("input data should be 2 integers when export type is %s", opts.Type)
				}

				// sort helper
				if opts.Input[0] > opts.Input[1] {
					opts.Input[0], opts.Input[1] = opts.Input[1], opts.Input[0]
				}
			case chat.ExportTypeLast:
				if len(opts.Input) != 1 {
					return fmt.Errorf("input data should be 1 integer when export type is %s", opts.Type)
				}
			default:
				return fmt.Errorf("unknown export type: %s", opts.Type)
			}

			// set default filters
			for _, filter := range chat.Filters {
				if opts.Filter[filter] == "" {
					opts.Filter[filter] = ".*"
				}
			}

			return chat.Export(logger.Named(cmd.Context(), "export"), &opts)
		},
	}

	const (
		_type = "type"
		input = "input"
	)

	utils.Cmd.StringEnumFlag(cmd, &opts.Type, _type, "T", chat.ExportTypeTime, []string{chat.ExportTypeTime, chat.ExportTypeID, chat.ExportTypeLast}, "export type. time: timestamp range, id: message id range, last: last N messages")
	cmd.Flags().StringVarP(&opts.Chat, "chat", "c", "", "chat id or domain")
	cmd.Flags().IntSliceVarP(&opts.Input, input, "i", []int{}, "input data, depends on export type")
	cmd.Flags().StringToStringVarP(&opts.Filter, "filter", "f", map[string]string{}, "only export media files that match the filter (regex). Default to all. Options: "+strings.Join(chat.Filters, ", "))
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "tdl-export.json", "output JSON file path")
	cmd.Flags().BoolVar(&opts.WithContent, "with-content", false, "export with message content")

	// completion and validation
	_ = cmd.RegisterFlagCompletionFunc(input, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// if user has already input something, don't do anything
		if toComplete != "" {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		switch cmd.Flags().Lookup(_type).Value.String() {
		case chat.ExportTypeTime:
			return []string{"0,9999999"}, cobra.ShellCompDirectiveNoFileComp
		case chat.ExportTypeID:
			return []string{"0,9999999"}, cobra.ShellCompDirectiveNoFileComp
		case chat.ExportTypeLast:
			return []string{"100"}, cobra.ShellCompDirectiveNoFileComp
		default:
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
	})

	return cmd
}
