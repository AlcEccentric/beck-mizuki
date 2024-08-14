package service

import (
	"sync"
	"time"

	"github.com/alceccentric/beck-crawler/dao/bgm"
	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
	job "github.com/alceccentric/beck-crawler/orch/job"
)

const (
	subjectDateFormat = "2006-01-02"

	// TODO: make these configurable
	earliestSubjectDate = "1993-10-01"
	latestSubjectDate   = "1993-12-31"
)

type SubjectService struct {
	bgmClient *bgm.BgmApiClient
}

func NewSubjectService(bgmClient *bgm.BgmApiClient) *SubjectService {
	return &SubjectService{
		bgmClient: bgmClient,
	}
}

func (svc *SubjectService) GetSubjectProducer(numOfSubjectProducers int) func(put func(*job.ColdStartOrchJob)) error {
	startDate, endDate := svc.getSubjectDateRange()
	daysPerProduer := int(endDate.Sub(startDate).Hours()/24) / numOfSubjectProducers

	return func(put func(*job.ColdStartOrchJob)) error {
		var wg sync.WaitGroup
		for i := 0; i < numOfSubjectProducers; i++ {
			index := i
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				curStartDate := startDate.Add(time.Duration(index*daysPerProduer) * 24 * time.Hour)
				curEndDate := curStartDate.Add(time.Duration(daysPerProduer) * 24 * time.Hour)

				if curEndDate.After(endDate) {
					curEndDate = endDate
				}

				subjects := svc.bgmClient.GetSubjectSlice(bgmModel.SubjectSearchRequest{
					Tag:          []string{"日本动画"},
					Type:         []bgmModel.SubjectType{bgmModel.Anime},
					AirDateRange: [2]time.Time{curStartDate, curEndDate},
					RatingRange:  [2]float32{0, 10},
				}, index)

				j := &job.ColdStartOrchJob{
					Subjects: subjects,
				}
				put(j)
			}(index)
		}
		wg.Wait()
		return nil
	}
}

func (svc *SubjectService) getSubjectDateRange() (startDate time.Time, endDate time.Time) {
	if sd, err := time.Parse(subjectDateFormat, earliestSubjectDate); err != nil {
		panic(err)
	} else {
		startDate = sd
	}

	if ed, err := time.Parse(subjectDateFormat, latestSubjectDate); err != nil {
		panic(err)
	} else {
		endDate = ed
	}
	return
}
