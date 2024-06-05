package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	DiscordToken string
)

var rootCmd = &cobra.Command{
	Use:   "infoe",
	Short: "main CLI for root command",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		if DiscordToken == "" {
			log.Fatal("--discord-token must not be empty")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&DiscordToken, "discord-token", "", "discord api token")
}
