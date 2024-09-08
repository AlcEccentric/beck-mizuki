package service

import (
	"sync"
	"time"

	dao "github.com/alceccentric/beck-crawler/dao"
	model "github.com/alceccentric/beck-crawler/model"
	job "github.com/alceccentric/beck-crawler/model/job"
	util "github.com/alceccentric/beck-crawler/util"
	"github.com/rs/zerolog/log"
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

				log.Info().Msgf("Trying to get subjects released between curStartDate: %s and curEndDate: %s", curStartDate, curEndDate)

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
					log.Info().Msgf("Got %d subjects between curStartDate: %s and curEndDate: %s", len(subjects), curStartDate, curEndDate)
					j := &job.ColdStartOrchJob{
						Subjects: subjects,
					}
					put(j)
				} else {
					log.Error().Msgf("Error getting subjects: %v. curStartDate: %s and curEndDate: %s. Skipping...", err, curStartDate, curEndDate)
				}

			}(index)
		}
		wg.Wait()
		return nil
	}
}

func (svc *SubjectService) getSubjectDateRange() (startDate time.Time, endDate time.Time) {
	if sd, err := time.Parse(util.SubjectDateFormat, util.EarliestSubjectDate); err != nil {
		log.Fatal().Err(err).Msgf("Error parsing earliest subject date %s", util.EarliestSubjectDate)
	} else {
		startDate = sd
	}

	if ed, err := time.Parse(util.SubjectDateFormat, util.LatestSubjectDate); err != nil {
		log.Fatal().Err(err).Msgf("Error parsing latest subject date %s", util.LatestSubjectDate)
	} else {
		endDate = ed
	}
	return
}
