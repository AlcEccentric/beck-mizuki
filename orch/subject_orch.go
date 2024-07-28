package orch

import (
	"fmt"
	"sync"
	"time"

	bgm "github.com/alceccentric/beck-crawler/dao/bgm"
	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
	"github.com/google/go-pipeline/pkg/pipeline"
)

const (
	earliestSubjectDate = "2020-01-01"
	latestSubjectDate   = "2020-12-31"
	dateFormat          = "2006-01-02"
)

type job struct {
	subjects []bgmModel.Subject
	users    []bgmModel.User
}

type SubjectOrchestrator struct {
	bgmClient                *bgm.BgmApiClient
	numOfSubjectProducers    int
	numOfUserProducers       int
	numOfCollectionProducers int
}

func NewSubjectOrchestrator(bgmClient *bgm.BgmApiClient, numOfCollectionProducers, numOfUserProducers, numOfSubjectProducers int) *SubjectOrchestrator {
	return &SubjectOrchestrator{
		bgmClient:                bgmClient,
		numOfSubjectProducers:    numOfSubjectProducers,
		numOfUserProducers:       numOfUserProducers,
		numOfCollectionProducers: numOfCollectionProducers,
	}
}

func (orch *SubjectOrchestrator) Run() {
	subjectProducerFn := orch.getSubjectProducer()
	userProducerFn := orch.getUserProducer()

	subjectProducer := pipeline.NewProducer(
		subjectProducerFn,
		pipeline.Name("job production"),
	)

	userProducer := pipeline.NewStage(
		userProducerFn,
		pipeline.Name("job handling"),
		pipeline.Concurrency(uint(orch.numOfUserProducers)),
	)

	if err := pipeline.Do(
		subjectProducer,
		userProducer,
	); err != nil {
		fmt.Printf("Do() failed: %s", err)
	}
}

func (orch *SubjectOrchestrator) getSubjectProducer() func(put func(*job)) error {
	startDate, endDate := getDates()
	daysPerProduer := int(endDate.Sub(startDate).Hours()/24) / orch.numOfSubjectProducers

	return func(put func(*job)) error {
		var wg sync.WaitGroup
		for i := 0; i < orch.numOfSubjectProducers; i++ {
			index := i
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				curStartDate := startDate.Add(time.Duration(index*daysPerProduer) * 24 * time.Hour)
				curEndDate := curStartDate.Add(time.Duration(daysPerProduer) * 24 * time.Hour)
				fmt.Printf("Producer: %d, start date: %s, end date: %s\n", index, curStartDate, curEndDate)

				if curEndDate.After(endDate) {
					curEndDate = endDate
				}

				subjects := orch.bgmClient.GetSubjectSlice(bgmModel.SubjectSearchRequest{
					Tag:          []string{"日本动画"},
					Type:         []bgmModel.SubjectType{bgmModel.Anime},
					AirDateRange: [2]time.Time{curStartDate, curEndDate},
					RatingRange:  [2]float32{0, 10},
				}, index)

				j := &job{
					subjects: subjects,
				}
				put(j)
			}(index)
		}
		wg.Wait()
		return nil
	}
}

func (orch *SubjectOrchestrator) getUserProducer() func(in *job) (*job, error) {
	return func(in *job) (*job, error) {
		for _, subject := range in.subjects {
			fmt.Printf("Received subject id: %s, name: %s\n", subject.Id, subject.Name)
		}
		return in, nil
	}
}

func getDates() (startDate time.Time, endDate time.Time) {
	if sd, err := time.Parse(dateFormat, earliestSubjectDate); err != nil {
		panic(err)
	} else {
		startDate = sd
	}

	if ed, err := time.Parse(dateFormat, latestSubjectDate); err != nil {
		panic(err)
	} else {
		endDate = ed
	}
	return
}
