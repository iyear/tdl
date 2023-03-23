package chat

import (
	"fmt"
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/spf13/cobra"
	"math"
	"strings"
)

var expOpts = &chat.ExportOptions{}

var cmdExport = &cobra.Command{
	Use:   "export",
	Short: "export messages from (protected) chat for download",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch expOpts.Type {
		case chat.ExportTypeTime, chat.ExportTypeID:
			// set default value
			switch len(expOpts.Input) {
			case 0:
				expOpts.Input = []int{0, math.MaxInt}
			case 1:
				expOpts.Input = append(expOpts.Input, math.MaxInt)
			}

			if len(expOpts.Input) != 2 {
				return fmt.Errorf("input data should be 2 integers when export type is %s", expOpts.Type)
			}

			// sort helper
			if expOpts.Input[0] > expOpts.Input[1] {
				expOpts.Input[0], expOpts.Input[1] = expOpts.Input[1], expOpts.Input[0]
			}
		case chat.ExportTypeLast:
			if len(expOpts.Input) != 1 {
				return fmt.Errorf("input data should be 1 integer when export type is %s", expOpts.Type)
			}
		default:
			return fmt.Errorf("unknown export type: %s", expOpts.Type)
		}

		// set default filters
		for _, filter := range chat.Filters {
			if expOpts.Filter[filter] == "" {
				expOpts.Filter[filter] = ".*"
			}
		}

		return chat.Export(logger.Named(cmd.Context(), "export"), expOpts)
	},
}

func init() {
	utils.Cmd.StringEnumFlag(cmdExport, &expOpts.Type, "type", "T", chat.ExportTypeTime, []string{chat.ExportTypeTime, chat.ExportTypeID, chat.ExportTypeLast}, "export type. time: timestamp range, id: message id range, last: last N messages")
	cmdExport.Flags().StringVarP(&expOpts.Chat, "chat", "c", "", "chat id or domain")
	cmdExport.Flags().IntSliceVarP(&expOpts.Input, "input", "i", []int{}, "input data, depends on export type")
	cmdExport.Flags().StringToStringVarP(&expOpts.Filter, "filter", "f", map[string]string{}, "only export media files that match the filter (regex). Default to all. Options: "+strings.Join(chat.Filters, ", "))
	cmdExport.Flags().StringVarP(&expOpts.Output, "output", "o", "tdl-export.json", "output JSON file path")
}
