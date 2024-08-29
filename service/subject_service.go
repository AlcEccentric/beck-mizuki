package service

import (
	"fmt"
	"sync"
	"time"

	dao "github.com/alceccentric/beck-crawler/dao"
	model "github.com/alceccentric/beck-crawler/model"
	job "github.com/alceccentric/beck-crawler/model/job"
)

const (
	subjectDateFormat = "2006-01-02"

	// TODO: make these configurable
	earliestSubjectDate = "2023-12-01"
	latestSubjectDate   = "2023-12-31"
)

type SubjectService struct {
	bgmClient *dao.BgmApiAccessor
}

func NewSubjectService(bgmClient *dao.BgmApiAccessor) *SubjectService {
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

				subjects, err := svc.bgmClient.GetSubjects(
					[]string{"日本动画"},
					[]model.SubjectType{model.Anime},
					[2]time.Time{curStartDate, curEndDate},
					[2]float32{0, 10},
				)

				if err == nil {
					j := &job.ColdStartOrchJob{
						Subjects: subjects,
					}
					put(j)
				} else {
					fmt.Printf("Error getting subjects: %v with curStartDate: %s and curEndDate: %s. Skipping...", err, curStartDate, curEndDate)
				}
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
