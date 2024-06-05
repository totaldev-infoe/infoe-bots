package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	mango "github.com/totaldev-infoe/infoe-bots/lib/mangostn"
)

var mangoCmd = &cobra.Command{
	Use:   "mango",
	Short: "Replies with pong if someone replies with ping",
	Long:  "N/A",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running mango function")
		mango.Mango(DiscordToken)
	},
}

func init() {
	rootCmd.AddCommand(mangoCmd)
}
