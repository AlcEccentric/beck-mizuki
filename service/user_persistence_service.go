package service

import (
	"fmt"

	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/helper"
	model "github.com/alceccentric/beck-crawler/model"
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
	for _, uid := range uids {
		// check if user meets criteria:
		isVIP, collections := helper.IsVip(uid, svc.bgmClient)
		if !isVIP {
			continue
		}

		// get user
		fmt.Printf("Found %d collections for user: %s\n", len(collections), uid)
		user, err := svc.getUser(uid)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to get user: %s (%w)", uid, err))
			continue
		}

		// insert user into db
		svc.konomiAccessor.InsertUser(user)
		for _, collection := range collections {
			// insert collection into db
			svc.konomiAccessor.InsertCollection(collection)
		}
	}
	return nil
}

func (svc *UserPersistenceService) getUser(uid string) (model.User, error) {
	latestCollectionTime, err := svc.bgmClient.GetCollectionTime(uid, 0)
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
