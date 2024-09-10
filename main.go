package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/model"
	. "github.com/alceccentric/beck-crawler/model/gen/beck-konomi/public/table"
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
	// bgmClient := dao.NewBgmApiAccessor()
	konomiAccessor := dao.NewCRKonomiAccessor()
	defer konomiAccessor.Disconnect()

	// Start execution
	// orch := orch.NewColdStartOrchestrator(bgmClient, konomiAccessor)
	// orch.Run(util.NumOfSubjectRetrievers, util.NumOfUserIdRetrievers, util.NumOfUserIdMergers)

	testUsers := make([]model.User, 0, 100)
	for i := 0; i < 100; i++ {
		testUsers = append(testUsers, model.User{
			ID:             strconv.Itoa(i + 1),
			Nickname:       "test" + strconv.Itoa(i+1),
			AvatarURL:      "test" + strconv.Itoa(i+1),
			LastActiveTime: time.Now(),
		})
	}
	konomiAccessor.BatchInsertUser(testUsers, 40)

	queriedUserIds, queryErr := konomiAccessor.GetUserIdsPaginated(10, 0)

	if queryErr == nil {
		for _, uid := range queriedUserIds {
			fmt.Println("UID: " + uid)
		}
	} else {
		fmt.Println(queryErr)
	}

	rn, countErr := konomiAccessor.GetRowCount(BgmUser)
	if countErr == nil {
		fmt.Println("Row count: " + strconv.Itoa(rn))
	}
}
