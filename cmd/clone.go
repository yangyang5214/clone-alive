package cmd

import (
	"os"

	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/internal"
	"github.com/yangyang5214/clone-alive/pkg/types"

	"github.com/spf13/cobra"
)

var (
	option = &types.Options{}
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a website",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			if err := cmd.Usage(); err != nil {
				gologger.Error().Msg(err.Error())
			}
			return
		}

		option.Url = args[0]
		r, err := internal.New(option)
		if err != nil {
			gologger.Error().Msgf("Created new crawler engine error: %s", err.Error())
			os.Exit(0)
		}
		r.Run()
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().BoolVarP(&option.Static, "static", "s", false, "use static as crawler engine")
	cloneCmd.Flags().BoolVarP(&option.Debug, "debug", "g", false, "debug ....")
	cloneCmd.Flags().Int8VarP(&option.MaxDepth, "depth", "d", 3, "max depth for crawler")
	cloneCmd.Flags().BoolVarP(&option.Append, "append", "a", false, "append model")
	cloneCmd.Flags().IntVarP(&option.MaxDuration, "duration", "u", 60*60*3, "max duration for crawler. default set 3h")
	cloneCmd.Flags().IntVarP(&option.Concurrent, "concurrent", "c", 5, "the number of concurrent crawling goroutines")
	cloneCmd.Flags().IntVarP(&option.Timeout, "timeout", "t", 30, "set timeout")
	cloneCmd.Flags().StringVarP(&option.Proxy, "proxy", "p", "", "set http proxy")
}
