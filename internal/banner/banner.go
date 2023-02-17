package banner

import (
	"github.com/projectdiscovery/gologger"
)

var (
	Version     = "v0.1.5"
	ProjectName = "clone-alive"
	//https://springhow.com/spring-boot-banner-generator/
	banner = "       _                          _ _\n   ___| | ___  _ __   ___    __ _| (_)_   _____\n  / __| |/ _ \\| '_ \\ / _ \\  / _` | | \\ \\ / / _ \\\n | (__| | (_) | | | |  __/ | (_| | | |\\ V /  __/\n  \\___|_|\\___/|_| |_|\\___|  \\__,_|_|_| \\_/ \\___|"
)

// ShowBanner is used to show the banner to the user
func ShowBanner() {
	gologger.Print().Msgf("%s, %s\n", banner, Version)
	gologger.Print().Msgf("\t\n%s\n\n", ProjectName)
}
