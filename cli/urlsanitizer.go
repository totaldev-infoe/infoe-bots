package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/totaldev-infoe/infoe-bots/lib/discord"
	"github.com/totaldev-infoe/infoe-bots/lib/urlsanitizer"
)

var urlSanitizerCmd = &cobra.Command{
	Use:   "urlsanitizer",
	Short: "Removes tracking parameters from URLs and warns about potentially unsafe links",
	Long:  "Monitors messages for URLs with tracking parameters (like Instagram's igsh) and reposts them without the tracking data. Also warns about potentially unsafe TikTok links.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running URL sanitizer")
		discord.Call(DiscordToken, urlsanitizer.HandleMessage)
	},
}

func init() {
	rootCmd.AddCommand(urlSanitizerCmd)
}
