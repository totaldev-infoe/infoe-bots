package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/totaldev-infoe/infoe-bots/lib/newegg"
)

var neweggLookupCmd = &cobra.Command{
	Use:   "newegg_lookup",
	Short: "Looks up RTX graphics card listings on Newegg",
	Long:  "Searches Newegg for RTX graphics card listings (5070, 5080, 5090) and reports their availability and prices",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Newegg RTX card lookup")
		newegg.LookupRTX(DiscordToken)
	},
}

func init() {
	rootCmd.AddCommand(neweggLookupCmd)
}
