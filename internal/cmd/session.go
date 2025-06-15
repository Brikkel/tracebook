package cmd

import (
	"log"

	"github.com/Brikkel/tracebook/internal/config"
	"github.com/Brikkel/tracebook/internal/session"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session [name]",
	Short: "Start a new documentation session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		session.StartPTYSession(args[0], cfg)
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}
