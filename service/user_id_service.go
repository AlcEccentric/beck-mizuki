package service

import (
	"sync"

	model "github.com/alceccentric/beck-crawler/model"
	orchJob "github.com/alceccentric/beck-crawler/model/job"
	"github.com/alceccentric/beck-crawler/scraper"
	"github.com/rs/zerolog/log"
)

type UserIdService struct {
}

func NewUserIdService() *UserIdService {
	return &UserIdService{}
}

func (orch *UserIdService) GetUserIdCollector(coldStartIntervalInDays int) func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error) {
	return func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error) {
		log.Info().Msgf("Retrieving ids for users who completed some works in the last %d days for %d subjects", coldStartIntervalInDays, len(in.Subjects))
		subjectUserScraper := scraper.NewSubjectUserScraper(coldStartIntervalInDays, len(in.Subjects))

		var wg sync.WaitGroup
		for _, subject := range in.Subjects {
			wg.Add(1)
			go func(subject model.Subject) {
				defer wg.Done()
				subjectUserScraper.Crawl(subject.Id)
			}(subject)
		}

		go func() {
			wg.Wait()
			subjectUserScraper.CloseUidChan()
		}()

		in.UserIds = subjectUserScraper.CollectUids()

		return in, nil
	}
}

func (orch *UserIdService) GetUserIdMerger() (func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error), map[string]struct{}) {
	userIdSet := make(map[string]struct{})

	userMergerFn := func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error) {
		log.Info().Msgf("Merging fetched user ids into one list (only keep unique user ids)")
		for _, userId := range in.UserIds {
			userIdSet[userId] = struct{}{}
		}
		return in, nil
	}

	return userMergerFn, userIdSet
}
