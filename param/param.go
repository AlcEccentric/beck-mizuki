package param

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/AlcEccentric/beck-mizuki/util"
	"github.com/rs/zerolog/log"
)

type Params struct {
	Mode                    ExecutionMode
	ColdStartIntervalInDays int
}

func GetParams() (params Params) {
	var modeStr string
	flag.StringVar(&modeStr, "mode", "", "mode: "+ColdStartMode.String()+" or "+RegularUpdateMode.String())
	flag.Parse()
	log.Info().Msgf("Retrieving CrawlerMode from flag arg string: %s", modeStr)

	return Params{
		Mode:                    getMode(modeStr),
		ColdStartIntervalInDays: getColdStartIntervalInDays(),
	}
}

func getColdStartIntervalInDays() int {
	coldStartIntervalInDays := os.Getenv("COLD_START_INTERVAL_IN_DAYS")
	if coldStartIntervalInDays == "" {
		log.Warn().Msg("COLD_START_INTERVAL_IN_DAYS environment variable is not set")
		return util.ColdStartIntervalInDays
	} else {
		coldStartIntervalInDaysInt, err := strconv.Atoi(coldStartIntervalInDays)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to parse COLD_START_INTERVAL_IN_DAYS %s", coldStartIntervalInDays)
		}
		return coldStartIntervalInDaysInt
	}
}

func getMode(modeStr string) ExecutionMode {
	if mode, err := CrawlerModeFromString(modeStr); err == nil {
		log.Info().Msgf("Retrieving CrawlerMode from mode arg string: %s", modeStr)
		return mode
	} else {
		log.Info().Msgf("Retrieving CrawlerMode from env var LAUNCH_DATE")
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
