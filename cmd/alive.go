package cmd

import (
	"github.com/projectdiscovery/gologger"
	"github.com/spf13/cobra"
)

// aliveCmd represents the alive command
var aliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Deploy as a honeypot",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gologger.Info().Msg("alive called")
	},
}

func init() {
	rootCmd.AddCommand(aliveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aliveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aliveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
