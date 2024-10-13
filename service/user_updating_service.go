package service

import (
	"math"
	"time"

	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/helper"
	model "github.com/alceccentric/beck-crawler/model"
	"github.com/alceccentric/beck-crawler/model/job"
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

func (svc *UserUpdatingService) GetUserUpdater() func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
	return func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
		activeUserIds := make([]string, 0)
		inactiveUserIds := make([]string, 0)
		log.Info().Msgf("Trying to update %d users", len(in.UserIds))
		for _, uid := range in.UserIds {
			// Get raw count
			rawWatchedCount, err := svc.bgmClient.GetCollectionCount(uid, model.Watched, model.Anime)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get raw watched count for user: %s. Skipping...", uid)
				continue
			}
			// check if user is still active (other check will always succeed for existing user, so we only check recent activity)
			isActive := helper.IsActive(svc.bgmClient, uid, rawWatchedCount)

			if isActive {
				log.Debug().Msgf("User %s is active", uid)
				activeUserIds = append(activeUserIds, uid)
			} else {
				log.Debug().Msgf("User %s is not active", uid)
				inactiveUserIds = append(inactiveUserIds, uid)
			}
		}

		svc.updateActiveUsers(activeUserIds)
		in.InactiveUserIds = inactiveUserIds
		return in, nil
	}
}

func (svc *UserUpdatingService) updateActiveUsers(uids []string) {
	for _, uid := range uids {
		user, getUserErr := svc.bgmClient.GetUser(uid)
		if getUserErr != nil {
			log.Error().Err(getUserErr).Msgf("Failed to get user: %s. Skipping...", uid)
			continue
		}
		daysSinceLastActive := math.Ceil(time.Since(user.LastActiveTime).Abs().Hours() / 24.0)
		collections, getCollectionsErr := svc.bgmClient.GetRecentCollections(uid, model.Watched, model.Anime, helper.AnimeFilter, int(daysSinceLastActive))
		if getCollectionsErr != nil {
			log.Error().Err(getCollectionsErr).Msgf("Failed to get recent watched collections for user: %s. Skipping...", uid)
			continue
		}

		svc.konomiAccessor.InsertUser(user)
		svc.konomiAccessor.BatchInsertCollection(collections, 100)
		log.Info().Msgf("Updated user: %s with %d collections", uid, len(collections))
	}

}
