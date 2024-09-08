package orch

import (
	"github.com/google/go-pipeline/pkg/pipeline"
	"github.com/rs/zerolog/log"

	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/service"
	"github.com/alceccentric/beck-crawler/util"
)

type ColdStartOrchestrator struct {
	bgmClient          *dao.BgmApiAccessor
	subjectSvc         *service.SubjectService
	userIdSvc          *service.UserIdService
	persistenceService *service.UserPersistenceService
}

func NewColdStartOrchestrator(bgmClient *dao.BgmApiAccessor, konomiAccessor *dao.KonomiAccessor) *ColdStartOrchestrator {
	return &ColdStartOrchestrator{
		bgmClient:          bgmClient,
		subjectSvc:         service.NewSubjectService(bgmClient),
		userIdSvc:          service.NewUserIdService(),
		persistenceService: service.NewUserPersistenceService(bgmClient, konomiAccessor),
	}
}

func (orch *ColdStartOrchestrator) Run(numOfSubjectProducers, numOfUserProducers, numOfUserIdMergers int) {
	log.Info().
		Int("numOfSubjectProducers", numOfSubjectProducers).
		Int("numOfUserProducers", numOfUserProducers).
		Int("numOfUserIdMergers", numOfUserIdMergers).
		Msg("Start cold start orchestrator")

	subjectProducerFn := orch.subjectSvc.GetSubjectProducer(numOfSubjectProducers)
	userProducerFn := orch.userIdSvc.GetUserIdCollector(util.ColdStartIntervalInDays)
	userMergerFn, userIdSet := orch.userIdSvc.GetUserIdMerger()

	subjectProducer := pipeline.NewProducer(
		subjectProducerFn,
		pipeline.Name("Retrieve subject data"),
	)

	userProducer := pipeline.NewStage(
		userProducerFn,
		pipeline.Name("Retrieve users that comment on subjects"),
		pipeline.Concurrency(uint(numOfUserProducers)),
	)

	userMerger := pipeline.NewStage(
		userMergerFn,
		pipeline.Name("Merge fetched user ids into one list (only keep unique user ids)"),
		pipeline.Concurrency(uint(numOfUserIdMergers)),
	)

	if err := pipeline.Do(
		subjectProducer,
		userProducer,
		userMerger,
	); err != nil {
		log.Error().Err(err).Msg("Failed to run cold startpipeline")
	} else {
		log.Info().Msgf("Fetched %d user ids", len(userIdSet))
		userIds := make([]string, 0, len(userIdSet))
		for uid := range userIdSet {
			userIds = append(userIds, uid)
		}
		orch.persistenceService.Persist(userIds)
	}
}
