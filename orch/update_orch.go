package orch

import (
	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/service"
	"github.com/google/go-pipeline/pkg/pipeline"
	"github.com/rs/zerolog/log"
)

type UpdateOrchestrator struct {
	bgmClient        *dao.BgmApiAccessor
	userIdReadingSvc *service.UserIdReadingService
	userUpdatingSvc  *service.UserUpdatingService
	userCleaningSvc  *service.UserCleaningService
}

func NewUpdateOrchestrator(bgmClient *dao.BgmApiAccessor, konomiAccessor dao.KonomiAccessor) *UpdateOrchestrator {
	return &UpdateOrchestrator{
		bgmClient:        bgmClient,
		userIdReadingSvc: service.NewUserIdReadingService(konomiAccessor),
		userUpdatingSvc:  service.NewUserUpdatingService(bgmClient, konomiAccessor),
		userCleaningSvc:  service.NewUserCleaningService(konomiAccessor),
	}
}

func (orch *UpdateOrchestrator) Run(numOfUserIdReaders, numOfCollectionUpdater, numOfDataCleaner int) {
	log.Info().
		Int("numOfUserIdRetrievers", numOfUserIdReaders).
		Int("numOfCollectionUpdater", numOfCollectionUpdater).
		Int("numOfDataCleaner", numOfDataCleaner).
		Msg("Start regular update orchestrator")

	// For user info each batch:
	// 1. perform VIP activity check
	// 2. Succeed:
	// 2.1. Upsert user info into db
	// 2.2. Insert new collections since last active time (also update last active time for the user)
	// 3. Fail:
	// 3.1. remove user & collections
	userIdReaderFn := orch.userIdReadingSvc.GetUserIdReader(numOfUserIdReaders)
	userUpdaterFn := orch.userUpdatingSvc.GetUserUpdater()
	userCleanerFn := orch.userCleaningSvc.GetUserCleaner()

	userIdReader := pipeline.NewProducer(
		userIdReaderFn,
		pipeline.Name("Read user ids from db"),
	)

	userUpdater := pipeline.NewStage(
		userUpdaterFn,
		pipeline.Name("Update info for active users and identify inactive users"),
	)

	userCleaner := pipeline.NewStage(
		userCleanerFn,
		pipeline.Name("Clean up inactive users"),
	)

	if err := pipeline.Do(
		userIdReader,
		userUpdater,
		userCleaner,
	); err != nil {
		log.Error().Err(err).Msg("Failed to run regular update pipeline")
	}
}
