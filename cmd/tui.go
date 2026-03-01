package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/iyear/tdl/app/tui"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
)

func NewTUI() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the Terminal User Interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize TUI model
			s := kv.From(cmd.Context()) // Get storage from context
			// We need to open the namespace
			ns := viper.GetString(consts.FlagNamespace)
			kvd, err := s.Open(ns)
			if err != nil {
				return fmt.Errorf("open kv storage: %w", err)
			}

			// Mute standard logger for TUI to prevent screen corruption
			muteLogger := zap.NewNop()
			cmd.SetContext(logctx.With(cmd.Context(), muteLogger))

			m := tui.NewModel(s, kvd, ns)

			// Start Bubble Tea program
			p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
			m.SetProgram(p) // Inject program reference
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error starting TUI: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
}
