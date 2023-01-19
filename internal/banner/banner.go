package banner

import (
	"github.com/projectdiscovery/gologger"
)

var (
	version     = "v0.0.4"
	ProjectName = "clone-alive"
	//https://springhow.com/spring-boot-banner-generator/
	banner = "       _                          _ _\n   ___| | ___  _ __   ___    __ _| (_)_   _____\n  / __| |/ _ \\| '_ \\ / _ \\  / _` | | \\ \\ / / _ \\\n | (__| | (_) | | | |  __/ | (_| | | |\\ V /  __/\n  \\___|_|\\___/|_| |_|\\___|  \\__,_|_|_| \\_/ \\___|"
)

// ShowBanner is used to show the banner to the user
func ShowBanner() {
	gologger.Print().Msgf("%s, %s\n", banner, version)
	gologger.Print().Msgf("\t\nclone-alive\n\n")
}
