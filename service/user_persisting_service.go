package service

import (
	"fmt"
	"math"
	"time"

	dao "github.com/AlcEccentric/beck-mizuki/dao"
	"github.com/AlcEccentric/beck-mizuki/helper"
	model "github.com/AlcEccentric/beck-mizuki/model"
	"github.com/rs/zerolog/log"
)

const (
	collectionInsertBatchSize = 50
)

type UserPersistingService struct {
	bgmClient      *dao.BgmApiAccessor
	konomiAccessor dao.KonomiAccessor
}

func NewUserPersistenceService(bgmClinet *dao.BgmApiAccessor, konomiAccessor dao.KonomiAccessor) *UserPersistingService {
	return &UserPersistingService{
		bgmClient:      bgmClinet,
		konomiAccessor: konomiAccessor,
	}
}

func (svc *UserPersistingService) Persist(uids []string) {
	log.Info().Msgf("Trying to persist %d users", len(uids))
	persistedUserCnt := 0
	for _, uid := range uids {
		// check if user meets criteria:
		user, getUserErr := svc.konomiAccessor.GetUser(uid)
		if getUserErr != nil {
			// user not found in db, meaning it's a new user
			isVIP, watchedCollections := helper.IsVip(uid, svc.bgmClient, svc.konomiAccessor)
			if isVIP {
				log.Info().Msgf("User %s is new and is a VIP, and will be persisted", uid)
				svc.insertUserWithQueriedCollections(uid, watchedCollections)
				persistedUserCnt++
			} else {
				log.Info().Msgf("User %s is new but is not a VIP", uid)
			}
		} else {
			// user already exists in db
			log.Info().Msgf("User %s already exists in db", uid)
			daysSinceLastActive := int(math.Ceil(time.Since(user.LastActiveTime).Abs().Hours() / 24.0))
			filteredWatched, err := svc.bgmClient.GetRecentCollections(uid, model.Watched, model.Anime, helper.AnimeFilter, daysSinceLastActive)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get filtered watched collections for user: %s. Skipping.", uid)
				return
			}
			log.Info().Msgf("Found %d filtered watched collections for user: %s in last %d days", len(filteredWatched), uid, daysSinceLastActive)

			if len(filteredWatched) > 0 {
				svc.insertUserWithQueriedCollections(uid, filteredWatched)
			}
			persistedUserCnt++
		}
	}
	log.Info().Msgf("In total, persisted %d users", persistedUserCnt)
}

func (svc *UserPersistingService) insertUserWithQueriedCollections(uid string, watchedCollections []model.Collection) {
	log.Info().Msgf("Found %d watched collections for user: %s", len(watchedCollections), uid)
	user, err := svc.getUser(uid)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get user: %s. Skipping...", uid)
		return
	}

	insertUserErr := svc.konomiAccessor.InsertUser(user)
	if insertUserErr == nil {
		svc.konomiAccessor.BatchInsertCollection(watchedCollections, collectionInsertBatchSize)
		log.Info().Msgf("Successfully persisted user: %s", uid)
	} else {
		log.Error().Err(insertUserErr).Msgf("Failed to persist user: %s. Skipping...", uid)
		return
	}
}

func (svc *UserPersistingService) getUser(uid string) (model.User, error) {
	latestCollectionTime, err := svc.bgmClient.GetCollectionTime(uid, 0, model.Watched, model.Anime)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get latest collection time for user: %s (%w)", uid, err)
	}

	user, err := svc.bgmClient.GetUser(uid)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get user: %s (%w)", uid, err)
	}
	return model.User{
		ID:             uid,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		LastActiveTime: latestCollectionTime,
	}, nil
}
