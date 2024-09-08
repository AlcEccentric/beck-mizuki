package main

import (
	"log"

	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/orch"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	// Set log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load env
	err := godotenv.Load("beck_mizuki.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Create dependencies (TODO: adopt DI if # of dependencies exceeds 3)
	bgmClient := dao.NewBgmApiAccessor()
	konomiAccessor := dao.NewKonomiAccessor()

	// Start execution
	orch := orch.NewColdStartOrchestrator(bgmClient, konomiAccessor)
	orch.Run(40, 1, 1)

	// Clean up
	konomiAccessor.Disconnect()
}
