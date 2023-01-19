package cmd

import (
	"github.com/projectdiscovery/gologger"
	"github.com/spf13/cobra"
	"github.com/yangyang5214/clone-alive/internal"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"os"
)

var (
	option = &types.Options{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clone-alive <url>",
	Short: "clone a website, then deploy as a honeypot ...",
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
	rootCmd.Flags().BoolVarP(&option.Headless, "headless", "a", true, "use chrome as crawler engine")
	rootCmd.Flags().BoolVarP(&option.Debug, "debug", "g", false, "debug ....")
	rootCmd.Flags().Int8VarP(&option.MaxDepth, "depth", "d", 3, "max depth for crawler")
	rootCmd.Flags().IntVarP(&option.MaxDuration, "duration", "u", 60*60*3, "max duration for crawler. default set 3h")
	rootCmd.Flags().IntVarP(&option.Concurrent, "concurrent", "c", 3, "the number of concurrent crawling goroutines")
	rootCmd.Flags().StringVarP(&option.Proxy, "proxy", "p", "", "set http proxy")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		gologger.Error().Msg(err.Error())
	}
}
