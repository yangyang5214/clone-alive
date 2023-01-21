package cmd

import (
	"github.com/projectdiscovery/gologger"
	"github.com/spf13/cobra"
	"github.com/yangyang5214/clone-alive/internal/banner"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "clone-alive",
	Short:   "clone a website, then deploy as a honeypot ...",
	Long:    ``,
	Version: banner.Version,
	Example: "clone-alive clone <url>\nclone-alive alive <dir>",
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		gologger.Error().Msg(err.Error())
	}
}
