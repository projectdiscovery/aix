package runner

import (
	"github.com/projectdiscovery/gologger"
	updateutils "github.com/projectdiscovery/utils/update"
)

const version = "v0.0.1"

var banner = (`
                         _  __
   ____ ___  ____ _____ | |/ /
  / __ '__ \/ __ '/ __ \|   / 
 / / / / / / /_/ / / / /   |  
/_/ /_/ /_/\__,_/_/ /_/_/|_|   Powered by OpenAI				  
`)

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")
}

// GetUpdateCallback returns a callback function that updates aix
func GetUpdateCallback() func() {
	return func() {
		showBanner()
		updateutils.GetUpdateToolCallback("aix", version)()
	}
}
