package service

import (
	"fmt"

	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/helper"
	model "github.com/alceccentric/beck-crawler/model"
	"github.com/rs/zerolog/log"
)

type UserPersistenceService struct {
	bgmClient      *dao.BgmApiAccessor
	konomiAccessor *dao.KonomiAccessor
}

func NewUserPersistenceService(bgmClinet *dao.BgmApiAccessor, konomiAccessor *dao.KonomiAccessor) *UserPersistenceService {
	return &UserPersistenceService{
		bgmClient:      bgmClinet,
		konomiAccessor: konomiAccessor,
	}
}

func (svc *UserPersistenceService) Persist(uids []string) error {
	log.Info().Msgf("Trying to persist %d users", len(uids))
	persistedUserCnt := 0
	for _, uid := range uids {
		// check if user meets criteria:
		isVIP, watchedCollections := helper.IsVip(uid, svc.bgmClient)
		if !isVIP {
			continue
		}

		log.Info().Msgf("Found %d watched collections for user: %s", len(watchedCollections), uid)
		user, err := svc.getUser(uid)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get user: %s. Skipping...", uid)
			continue
		}

		svc.konomiAccessor.InsertUser(user)
		for _, collection := range watchedCollections {
			svc.konomiAccessor.InsertCollection(collection)
		}
		log.Info().Msgf("Successfully persisted user: %s", uid)
		persistedUserCnt++
	}
	log.Info().Msgf("At the end, persisted %d users", persistedUserCnt)
	return nil
}

func (svc *UserPersistenceService) getUser(uid string) (model.User, error) {
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
