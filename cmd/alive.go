package cmd

import (
	"os"
	"path/filepath"

	"github.com/projectdiscovery/gologger"
	"github.com/spf13/cobra"
	"github.com/yangyang5214/clone-alive/pkg/alive"
	"github.com/yangyang5214/clone-alive/pkg/output"
	"github.com/yangyang5214/clone-alive/pkg/types"
)

var aliveOption types.AliveOption

// aliveCmd represents the alive command
var aliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Deploy as a honeypot",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			if err := cmd.Usage(); err != nil {
				gologger.Error().Msg(err.Error())
			}
			return
		}
		aliveOption.HomeDir = args[0]
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
	aliveCmd.Flags().IntVarP(&aliveOption.Port, "port", "p", 8001, "port for server")
	aliveCmd.Flags().BoolVarP(&aliveOption.Debug, "debug", "b", false, "debug log level")
	aliveCmd.Flags().StringVarP(&aliveOption.HomeDir, "home-dir", "d", "", "static file dir")
	aliveCmd.Flags().StringVarP(&aliveOption.VerifyCodePath, "verify_code", "", "config/verify_code", "verify_code file path")
	aliveCmd.Flags().StringVarP(&aliveOption.CustomRulePath, "custom_rule", "", "config/custom_rule", "custom_rule file path")
}
