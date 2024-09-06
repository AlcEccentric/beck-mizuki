package orch

import (
	"fmt"

	"github.com/google/go-pipeline/pkg/pipeline"

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
		pipeline.Name("User collection data persistion"),
		pipeline.Concurrency(uint(numOfUserIdMergers)),
	)

	if err := pipeline.Do(
		subjectProducer,
		userProducer,
		userMerger,
	); err != nil {
		fmt.Printf("Do() failed: %s", err)
	} else {
		fmt.Printf("Received %d users\n", len(userIdSet))
		userIds := make([]string, 0, len(userIdSet))
		for uid := range userIdSet {
			userIds = append(userIds, uid)
		}
		orch.persistenceService.Persist(userIds)
	}
}
