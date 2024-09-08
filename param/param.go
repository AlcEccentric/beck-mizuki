package param

import (
	"flag"
	"os"
	"time"

	"github.com/alceccentric/beck-crawler/util"
	"github.com/rs/zerolog/log"
)

type Params struct {
	Mode ExecutionMode
}

func GetParams() (params Params) {
	var modeStr string
	flag.StringVar(&modeStr, "mode", "", "mode: "+ColdStartMode.String()+" or "+RegularUpdateMode.String())
	flag.Parse()

	return Params{
		Mode: getMode(modeStr),
	}
}

func getMode(modeStr string) ExecutionMode {
	if mode, err := CrawlerModeFromString(modeStr); err == nil {
		log.Info().Msgf("Retrieving CrawlerMode from mode arg: %s", modeStr)
		return mode
	} else {
		log.Info().Msgf("Retrieving CrawlerMode per days since launch date")
		launchDateStr := os.Getenv("LAUNCH_DATE")
		if launchDateStr == "" {
			log.Fatal().Msg("LAUNCH_DATE environment variable is not set")
		}

		launchDate, err := time.Parse(util.LaunchDateFormat, launchDateStr)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to parse LAUNCH_DATE %s", launchDateStr)
		}

		daysSinceLaunch := int(time.Since(launchDate).Hours() / 24)
		log.Info().Msgf("It has been %d days since launch", daysSinceLaunch)

		mode := DirectlyExitMode
		if daysSinceLaunch%util.ColdStartIntervalInDays == 0 {
			mode = ColdStartMode
		} else if daysSinceLaunch%util.RegularUpdateIntervalInDays == 0 {
			mode = RegularUpdateMode
		}
		log.Info().Msgf("Running in mode: %s", mode.String())
		return mode
	}
}
