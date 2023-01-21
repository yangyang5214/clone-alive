package cmd

import (
	"github.com/projectdiscovery/gologger"
	"github.com/spf13/cobra"
	"github.com/yangyang5214/clone-alive/pkg/alive"
	"github.com/yangyang5214/clone-alive/pkg/output"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"os"
	"path/filepath"
)

var aliveOption types.AliveOption

// aliveCmd represents the alive command
var aliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Deploy as a honeypot",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		aliveOption.RouteFile = filepath.Join(aliveOption.HomeDir, output.RouterFile)
		a := alive.New(aliveOption)
		err := a.Run()
		if err != nil {
			gologger.Error().Msg(err.Error())
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(aliveCmd)
	aliveCmd.Flags().IntVarP(&aliveOption.Port, "port", "p", 8080, "default port for web server")
	aliveCmd.Flags().BoolVarP(&aliveOption.Debug, "debug", "b", false, "debug model for gin")
	aliveCmd.Flags().StringVarP(&aliveOption.HomeDir, "home-dir", "d", "", "static file dir")
	_ = aliveCmd.MarkFlagRequired("home-dir")
}
