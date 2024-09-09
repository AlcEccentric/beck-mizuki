package service

import (
	"os"
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

func (svc *SubjectService) GetSubjectRetriever(numOfSubjectRetrievers int) func(put func(*job.ColdStartOrchJob)) error {
	startDate, endDate := svc.getSubjectDateRange()
	dateRanges := divideDateRanges(startDate, endDate, numOfSubjectRetrievers)

	return func(put func(*job.ColdStartOrchJob)) error {
		var wg sync.WaitGroup
		for i := 0; i < numOfSubjectRetrievers; i++ {
			index := i
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				curDateRange := dateRanges[index]
				curStartDate := curDateRange[0]
				curEndDate := curDateRange[1]

				if curStartDate.Before(curEndDate) {
					log.Info().Msgf("Trying to get subjects released between curStartDate: %s and curEndDate: %s", curStartDate, curEndDate)

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
				}
			}(index)
		}
		wg.Wait()
		return nil
	}
}

func (svc *SubjectService) getSubjectDateRange() (startDate time.Time, endDate time.Time) {
	// get earliest subject date
	startDateStr := os.Getenv("START_SUBJECT_DATE")
	if startDateStr == "" {
		log.Fatal().Msg("START_SUBJECT_DATE environment variable is not set")
	}

	if sd, err := time.Parse(util.SubjectDateFormat, startDateStr); err != nil {
		log.Fatal().Err(err).Msgf("Error parsing start subject date %s", startDateStr)
	} else {
		startDate = sd
	}

	// get latest subject date
	endDateStr := os.Getenv("END_SUBJECT_DATE")
	if endDateStr == "" {
		endDate = time.Now()
	} else {
		if ed, err := time.Parse(util.SubjectDateFormat, endDateStr); err != nil {
			log.Fatal().Err(err).Msgf("Error parsing end subject date %s", endDateStr)
		} else {
			endDate = ed
		}
	}

	log.Info().Msgf("Subject fetching start date: %s end date: %s", startDate, endDate)
	return
}

func divideDateRanges(startDate, endDate time.Time, numWorkers int) map[int][2]time.Time {
	daysBetween := int(endDate.Sub(startDate).Hours() / 24)
	baseRange := daysBetween / numWorkers
	remainder := daysBetween % numWorkers

	dateRanges := make(map[int][2]time.Time)

	currentStart := startDate
	for i := 0; i < numWorkers; i++ {
		rangeSize := baseRange
		if i < remainder {
			rangeSize++
		}
		currentEnd := currentStart.AddDate(0, 0, rangeSize)
		if currentEnd.After(endDate) {
			currentEnd = endDate
		}
		dateRanges[i] = [2]time.Time{currentStart, currentEnd}
		currentStart = currentEnd
	}

	return dateRanges
}
