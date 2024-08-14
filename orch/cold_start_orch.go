package orch

import (
	"fmt"

	"github.com/google/go-pipeline/pkg/pipeline"

	bgm "github.com/alceccentric/beck-crawler/dao/bgm"
	"github.com/alceccentric/beck-crawler/service"
)

const (
	dateFormat = "2006-01-02"

	// TODO: make these configurable
	earliestSubjectDate         = "1993-10-01"
	latestSubjectDate           = "1993-12-31"
	collectionTimeHorizonInDays = 30
)

type ColdStartOrchestrator struct {
	bgmClient                *bgm.BgmApiClient
	numOfSubjectProducers    int
	numOfUserProducers       int
	numOfCollectionProducers int
	subjectSvc               *service.SubjectService
	userIdSvc                *service.UserIdService
}

func NewColdStartOrchestrator(bgmClient *bgm.BgmApiClient, numOfCollectionProducers, numOfUserProducers, numOfSubjectProducers int) *ColdStartOrchestrator {
	return &ColdStartOrchestrator{
		bgmClient:                bgmClient,
		numOfSubjectProducers:    numOfSubjectProducers,
		numOfUserProducers:       numOfUserProducers,
		numOfCollectionProducers: numOfCollectionProducers,
		subjectSvc:               service.NewSubjectService(bgmClient),
		userIdSvc:                service.NewUserIdService(),
	}
}

func (orch *ColdStartOrchestrator) Run() {
	subjectProducerFn := orch.subjectSvc.GetSubjectProducer(orch.numOfSubjectProducers)
	userProducerFn := orch.userIdSvc.GetUserIdCollector(orch.numOfUserProducers)
	userMergerFn, userIdSet := orch.userIdSvc.GetUserIdMerger()

	subjectProducer := pipeline.NewProducer(
		subjectProducerFn,
		pipeline.Name("Retrieve subject data"),
	)

	userProducer := pipeline.NewStage(
		userProducerFn,
		pipeline.Name("Retrieve users that comment on subjects"),
		pipeline.Concurrency(uint(orch.numOfUserProducers)),
	)

	userMerger := pipeline.NewStage(
		userMergerFn,
		pipeline.Name("User collection data persistion"),
		pipeline.Concurrency(uint(orch.numOfUserProducers)),
	)

	if err := pipeline.Do(
		subjectProducer,
		userProducer,
		userMerger,
	); err != nil {
		fmt.Printf("Do() failed: %s", err)
	} else {
		fmt.Println(len(userIdSet))
	}
}
