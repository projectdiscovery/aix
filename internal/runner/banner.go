package runner

import (
	"github.com/projectdiscovery/gologger"
	updateutils "github.com/projectdiscovery/utils/update"
)

const version = "v0.0.5"

var banner = (`
   ___   _____  __
  / _ | /  _/ |/_/
 / __ |_/ /_>  < 
/_/ |_/___/_/|_|  Powered by OpenAI

   projectdiscovery.io		  
`)

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
}

// GetUpdateCallback returns a callback function that updates aix
func GetUpdateCallback() func() {
	return func() {
		showBanner()
		updateutils.GetUpdateToolCallback("aix", version)()
	}
}
