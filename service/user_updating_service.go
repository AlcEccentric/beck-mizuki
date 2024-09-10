package service

import (
	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/helper"
	model "github.com/alceccentric/beck-crawler/model"
	"github.com/alceccentric/beck-crawler/model/job"
	util "github.com/alceccentric/beck-crawler/util"
	"github.com/rs/zerolog/log"
)

type UserUpdatingService struct {
	bgmClient      *dao.BgmApiAccessor
	konomiAccessor dao.KonomiAccessor
}

func NewUserUpdatingService(
	bgmClient *dao.BgmApiAccessor,
	konomiAccessor dao.KonomiAccessor,
) *UserUpdatingService {
	return &UserUpdatingService{
		bgmClient:      bgmClient,
		konomiAccessor: konomiAccessor,
	}
}

func (s *UserUpdatingService) GetUserUpdater() func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
	return func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
		activeUserIds := make([]string, 0)
		inactiveUserIds := make([]string, 0)
		for _, uid := range in.UserIds {
			log.Info().Msgf("Inspecting user: %s", uid)

			// Get raw count
			rawWatchedCount, err := s.bgmClient.GetCollectionCount(uid, model.Watched, model.Anime)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get raw watched count for user: %s. Skipping...", uid)
				continue
			}
			// check if user is still active (other check will always succeed for existing user, so we only check recent activity)
			isActive := helper.IsActive(s.bgmClient, uid, rawWatchedCount)

			if isActive {
				activeUserIds = append(activeUserIds, uid)
			} else {
				inactiveUserIds = append(inactiveUserIds, uid)
			}
		}

		s.updateActiveUsers(activeUserIds)
		in.InactiveUserIds = inactiveUserIds
		return in, nil
	}
}

func (s *UserUpdatingService) updateActiveUsers(uids []string) {

	for _, uid := range uids {
		user, getUserErr := s.bgmClient.GetUser(uid)
		if getUserErr != nil {
			log.Error().Err(getUserErr).Msgf("Failed to get user: %s. Skipping...", uid)
			continue
		}

		collections, getCollectionsErr := s.bgmClient.GetRecentCollections(uid, model.Watched, model.Anime, helper.AnimeFilter, util.RegularUpdateIntervalInDays)
		if getCollectionsErr != nil {
			log.Error().Err(getCollectionsErr).Msgf("Failed to get recent watched collections for user: %s. Skipping...", uid)
			continue
		}

		s.konomiAccessor.InsertUser(user)
		s.konomiAccessor.BatchInsertCollection(collections, 100)
	}

}
