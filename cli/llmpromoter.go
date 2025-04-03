package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/totaldev-infoe/infoe-bots/lib/discord"
	"github.com/totaldev-infoe/infoe-bots/lib/llmpromoter"
)

var llmPromoterCmd = &cobra.Command{
	Use:   "llmpromoter",
	Short: "Promotes local LLMs when users complain about hosted services",
	Long:  "Monitors messages for complaints about hosted LLM services (like Claude, ChatGPT, etc.) being down or slow and suggests local LLM alternatives.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running LLM Promoter")
		discord.Call(DiscordToken, llmpromoter.HandleMessage, llmpromoter.HandleReactionAdd)
	},
}

func init() {
	rootCmd.AddCommand(llmPromoterCmd)
}
