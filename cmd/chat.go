package cmd

import (
	"fmt"
	"math"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/pkg/logger"
)

func NewChat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "A set of chat tools",
	}

	cmd.AddCommand(NewChatList(), NewChatExport(), NewChatUsers())

	return cmd
}

func NewChatList() *cobra.Command {
	var opts chat.ListOptions

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List your chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return chat.List(logger.Named(cmd.Context(), "ls"), opts)
		},
	}

	cmd.Flags().VarP(&opts.Output, "output", "o", fmt.Sprintf("output format: [%s]", strings.Join(chat.OutputNames(), ", ")))
	cmd.Flags().StringVarP(&opts.Filter, "filter", "f", "true", "filter chats by expression")

	return cmd
}

func NewChatExport() *cobra.Command {
	var opts chat.ExportOptions

	cmd := &cobra.Command{
		Use:   "export",
		Short: "export messages from (protected) chat for download",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch opts.Type {
			case chat.ExportTypeTime, chat.ExportTypeId:
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

			return chat.Export(logger.Named(cmd.Context(), "export"), &opts)
		},
	}

	const (
		_type = "type"
		_chat = "chat"
		input = "input"
	)

	cmd.Flags().VarP(&opts.Type, _type, "T", fmt.Sprintf("export type: [%s]", strings.Join(chat.ExportTypeNames(), ", ")))
	cmd.Flags().StringVarP(&opts.Chat, _chat, "c", "", "chat id or domain. If not specified, 'Saved Messages' will be used")

	// topic id and message id is the same field in tg.MessagesGetRepliesRequest
	cmd.Flags().IntVar(&opts.Thread, "topic", 0, "specify topic id")
	cmd.Flags().IntVar(&opts.Thread, "reply", 0, "specify channel post id")

	cmd.Flags().IntSliceVarP(&opts.Input, input, "i", []int{}, "input data, depends on export type")
	cmd.Flags().StringVarP(&opts.Filter, "filter", "f", "true", "filter messages by expression, defaults to match all messages. Specify '-' to see available fields")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "tdl-export.json", "output JSON file path")
	cmd.Flags().BoolVar(&opts.WithContent, "with-content", false, "export with message content")
	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "export raw message struct of Telegram MTProto API, useful for debugging")
	cmd.Flags().BoolVar(&opts.All, "all", false, "export all messages including non-media messages, but still affected by filter and type flag")

	// completion and validation
	_ = cmd.RegisterFlagCompletionFunc(input, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// if user has already input something, don't do anything
		if toComplete != "" {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		switch cmd.Flags().Lookup(_type).Value.String() {
		case chat.ExportTypeTime.String():
			return []string{"0,9999999"}, cobra.ShellCompDirectiveNoFileComp
		case chat.ExportTypeId.String():
			return []string{"0,9999999"}, cobra.ShellCompDirectiveNoFileComp
		case chat.ExportTypeLast.String():
			return []string{"100"}, cobra.ShellCompDirectiveNoFileComp
		default:
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
	})

	return cmd
}

func NewChatUsers() *cobra.Command {
	var opts chat.UsersOptions

	cmd := &cobra.Command{
		Use:   "users",
		Short: "export users from (protected) channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			return chat.Users(logger.Named(cmd.Context(), "users"), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "tdl-users.json", "output JSON file path")
	cmd.Flags().StringVarP(&opts.Chat, "chat", "c", "", "domain id (channels, supergroups, etc.)")
	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "export raw message struct of Telegram MTProto API, useful for debugging")
	return cmd
}
