package orch

import (
	"github.com/google/go-pipeline/pkg/pipeline"
	"github.com/rs/zerolog/log"

	dao "github.com/AlcEccentric/beck-mizuki/dao"
	"github.com/AlcEccentric/beck-mizuki/service"
)

type ColdStartOrchestrator struct {
	bgmClient          *dao.BgmApiAccessor
	subjectSvc         *service.SubjectService
	userIdSvc          *service.UserIdScrapingService
	persistenceService *service.UserPersistingService
}

func NewColdStartOrchestrator(bgmClient *dao.BgmApiAccessor, konomiAccessor dao.KonomiAccessor) *ColdStartOrchestrator {
	return &ColdStartOrchestrator{
		bgmClient:          bgmClient,
		subjectSvc:         service.NewSubjectService(bgmClient),
		userIdSvc:          service.NewUserIdScrapingService(),
		persistenceService: service.NewUserPersistenceService(bgmClient, konomiAccessor),
	}
}

func (orch *ColdStartOrchestrator) Run(numOfSubjectRetrievers, numOfUserIdRetrievers, numOfUserIdMergers, coldStartIntervalInDays int) {
	log.Info().
		Int("numOfSubjectRetrievers", numOfSubjectRetrievers).
		Int("numOfUserIdRetrievers", numOfUserIdRetrievers).
		Int("numOfUserIdMergers", numOfUserIdMergers).
		Int("coldStartIntervalInDays", coldStartIntervalInDays).
		Msg("Start cold start orchestrator")

	subjectRetrieverFn := orch.subjectSvc.GetSubjectRetriever(numOfSubjectRetrievers)
	userIdRetrieverFn := orch.userIdSvc.GetUserIdRetriever(coldStartIntervalInDays)
	userMergerFn, userIdSet := orch.userIdSvc.GetUserIdMerger()

	subjectRetriever := pipeline.NewProducer(
		subjectRetrieverFn,
		pipeline.Name("Retrieve subject data"),
	)

	userIdRetriever := pipeline.NewStage(
		userIdRetrieverFn,
		pipeline.Name("Retrieve users that comment on subjects"),
		pipeline.Concurrency(uint(numOfUserIdRetrievers)),
	)

	userMerger := pipeline.NewStage(
		userMergerFn,
		pipeline.Name("Merge fetched user ids into one list (only keep unique user ids)"),
		pipeline.Concurrency(uint(numOfUserIdMergers)),
	)

	if err := pipeline.Do(
		subjectRetriever,
		userIdRetriever,
		userMerger,
	); err != nil {
		log.Error().Err(err).Msg("Failed to run cold start pipeline")
	} else {
		log.Info().Msgf("Fetched %d user ids", len(userIdSet))
		userIds := make([]string, 0, len(userIdSet))
		for uid := range userIdSet {
			userIds = append(userIds, uid)
		}
		orch.persistenceService.Persist(userIds)
	}
}
