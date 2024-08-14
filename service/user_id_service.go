package service

import (
	"sync"

	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
	orchJob "github.com/alceccentric/beck-crawler/orch/job"
	"github.com/alceccentric/beck-crawler/scraper"
)

type UserIdService struct {
}

func NewUserIdService() *UserIdService {
	return &UserIdService{}
}

func (orch *UserIdService) GetUserIdCollector(collectionTimeHorizonInDays int) func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error) {
	return func(in *orchJob.ColdStartOrchJob) (*orchJob.ColdStartOrchJob, error) {
		subjectUserScraper := scraper.NewSubjectUserScraper(collectionTimeHorizonInDays, len(in.Subjects))

		var wg sync.WaitGroup
		for _, subject := range in.Subjects {
			wg.Add(1)
			go func(subject bgmModel.Subject) {
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
		for _, userId := range in.UserIds {
			userIdSet[userId] = struct{}{}
		}
		return in, nil
	}

	return userMergerFn, userIdSet
}
