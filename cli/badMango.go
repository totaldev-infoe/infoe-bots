package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	badMango "github.com/totaldev-infoe/infoe-bots/lib/mangosteen"
)

var badMangoCmd = &cobra.Command{
	Use:   "badMango",
	Short: "Discourages message deletion by mango",
	Long:  "N/A",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running badMango function")
		badMango.BadMango(DiscordToken)
	},
}

func init() {
	rootCmd.AddCommand(badMangoCmd)
}
