package main

import (
	"os"

	"github.com/AlcEccentric/beck-mizuki/dao"
	"github.com/AlcEccentric/beck-mizuki/orch"
	"github.com/AlcEccentric/beck-mizuki/param"
	"github.com/AlcEccentric/beck-mizuki/util"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load env
	err := godotenv.Load("beck_mizuki.env")
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Create dependencies (TODO: adopt DI if # of dependencies exceeds 3)
	bgmClient := dao.NewBgmApiAccessor()
	konomiAccessor := dao.NewCRKonomiAccessor()
	defer konomiAccessor.Disconnect()

	// Start execution
	params := param.GetParams()
	if params.Mode == param.ColdStartMode {
		orch := orch.NewColdStartOrchestrator(bgmClient, konomiAccessor)
		orch.Run(util.NumOfSubjectRetrievers, util.NumOfUserIdRetrievers, util.NumOfUserIdMergers, params.ColdStartIntervalInDays)
	} else if params.Mode == param.RegularUpdateMode {
		orch := orch.NewUpdateOrchestrator(bgmClient, konomiAccessor)
		orch.Run(util.NumOfUserIDReaders, util.NumOfUserUpdaters, util.NumOfUserCleaners)
	} else {
		log.Info().Msg("Not on a run date. Exiting...")
		os.Exit(0)
	}
}
