package service

import (
	"fmt"
	"time"

	dao "github.com/alceccentric/beck-crawler/dao"
	model "github.com/alceccentric/beck-crawler/model"
	"github.com/tidwall/gjson"
)

const (
	// No need to check recent activity in this class
	// as the SubUser_Scaper already does that when fetching users from collection page
	minActiveTimeInDays = 365
	minCollectionCount  = 300
)

var tagsToReject = map[string]struct{}{
	"国产":   {},
	"国产动画": {},
	"中国":   {},
	"欧美":   {},
	"童年":   {},
	"短片":   {},
	"PV":   {},
	"民工":   {},
	"MV":   {},
}

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
		if !svc.isVIP(uid) {
			continue
		}

		// look up latest collection time
		user, err := svc.getUser(uid)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to get user: %s (%w)", uid, err))
			continue
		}

		// look up all user collections
		collections, err := svc.bgmClient.GetCollections(uid, model.Watched, model.Anime, animeCollectionAcceptor)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to get collections for user: %s (%w)", uid, err))
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

func (svc *UserPersistenceService) isVIP(uid string) bool {
	// a valid vip user must meet the following criteria:
	// watched at least minWatchedAnimeCount anime
	// earliest collection is at least minActiveTimeInDays from today
	// inactive for at most maxInactiveTimeInDays
	collectedEnough, collectionCnt := svc.hasEnoughCollection(uid)
	if !collectedEnough {
		return false
	}

	if !svc.isLoyal(uid, collectionCnt) {
		return false
	}

	return true
}

func (svc *UserPersistenceService) hasEnoughCollection(uid string) (bool, int) {
	collectionCount, err := svc.bgmClient.GetCollectionCount(uid, model.Watched, model.Anime)

	if err != nil {
		fmt.Println(err)
		return false, -1
	}

	if collectionCount < minCollectionCount {
		fmt.Println(fmt.Errorf("user: %s has not collected %d subjects", uid, minCollectionCount))
		return false, collectionCount
	}
	return true, collectionCount
}

func (svc *UserPersistenceService) isLoyal(uid string, collected int) bool {
	earliestCollectionTime, err := svc.bgmClient.GetCollectionTime(uid, collected-1)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if time.Since(earliestCollectionTime) < time.Hour*24*minActiveTimeInDays {
		fmt.Println(fmt.Errorf("user: %s earliest collection time was under %d days from today", uid, minActiveTimeInDays))
		return false
	}
	return true
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

func animeCollectionAcceptor(animeCol gjson.Result) bool {
	tags := animeCol.Get("subject").Get("tags").Array()
	for _, tag := range tags {
		if _, ok := tagsToReject[tag.Get("name").String()]; ok {
			return false
		}
	}
	collectionTotal := animeCol.Get("subject").Get("collection_total").Int()
	// assuming a subject with too few collections are not generally available
	// meaning not watching it does not necessarily mean people are not interested in the work
	return collectionTotal >= minCollectionCount
}
