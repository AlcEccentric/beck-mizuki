package helper

import (
	"fmt"
	"time"

	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/model"
	"github.com/alceccentric/beck-crawler/util"
	"github.com/tidwall/gjson"
)

// Reject users with watched count less than T1WatchedCnt
// Reject users whose oldest watched collection was made more than last MinOldestWatchedAgeInDays days ago
// For users with watched count in [T1WatchedCnt, T2WatchedCnt), they should have at least 1 watched every T1IntervalDays in the past ActivityCheckDays
// For users with watched count in [T2WatchedCnt, T3WatchedCnt), they should have at least 1 watched every T2IntervalDays in the past ActivityCheckDays
// For users with watched count in [T3WatchedCnt, inf), they should have at least 1 collection every T3IntervalDays in the past ActivityCheckDays
// If user does not meet the criteria for watched, they should have at least MinWatchingCnt in the past ActivityCheckDays
// For each user, up to NonWatchedIntervalTolerance periods are allowed without a collection
// Reject users with less than MinFilteredWatchedCnt collections

var tagsToReject = map[string]struct{}{
	"国产":   {},
	"国产动画": {},
	"中国":   {},
	"欧美":   {},
	"美国":   {},
	"童年":   {},
	"短片":   {},
	"PV":   {},
	"民工":   {},
	"MV":   {},
}

func IsVip(uid string, bgmAPI *dao.BgmApiAccessor) (bool, []model.Collection) {
	// raw watched collection count filter
	rawWatchedCount, err := bgmAPI.GetCollectionCount(uid, model.Watched, model.Anime)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if rawWatchedCount < util.T1WatchedCnt {
		fmt.Println(fmt.Errorf("user: %s raw collection count was under %d", uid, util.T1WatchedCnt))
		return false, nil
	}

	// oldest collection filter
	earliestWatchedTime, err := bgmAPI.GetCollectionTime(uid, rawWatchedCount-1)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if time.Since(earliestWatchedTime) < time.Hour*24*util.MinOldestWatchedAgeInDays {
		fmt.Println(fmt.Errorf("user: %s earliest watched collection time was under %d days from today", uid, util.MinOldestWatchedAgeInDays))
		return false, nil
	}

	// leveled activity check
	watched, err := bgmAPI.GetRecentCollections(uid, model.Watched, model.Anime, animeFilter, util.ActivityCheckDays)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	active := true
	if rawWatchedCount < util.T2WatchedCnt {
		active = isActiveWatched(watched, util.T1IntervalDays) || isActiveWatching(bgmAPI, uid)
	} else if rawWatchedCount < util.T3WatchedCnt {
		active = isActiveWatched(watched, util.T2IntervalDays) || isActiveWatching(bgmAPI, uid)
	} else {
		active = isActiveWatched(watched, util.T3IntervalDays) || isActiveWatching(bgmAPI, uid)
	}

	if !active {
		fmt.Printf("user: %s with is not considered active\n", uid)
		return false, nil
	}

	// filtered watched count check
	filteredWatched, err := bgmAPI.GetCollections(uid, model.Watched, model.Anime, animeFilter)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if len(filteredWatched) < util.MinFilteredWatchedCnt {
		fmt.Printf("user: %s with %d filtered watched collections is under %d\n", uid, len(filteredWatched), util.MinFilteredWatchedCnt)
		return false, nil
	}

	// Return filtered watched to reduce the number of API calls
	return true, filteredWatched
}

func isActiveWatched(collections []model.Collection, intervalDays int) bool {
	lastIntervalIdx := -1
	toleranceCounter := util.NonWatchedIntervalTolerance
	for i := 0; i < len(collections); i++ {
		collectionTime, err := time.Parse(util.CollectionTimeFormat, collections[i].CollectedTime)
		if err != nil {
			panic(fmt.Errorf("failed to parse collection time: %s (%w)", collections[i].CollectedTime, err))
		}
		curIntervalIdx := int(time.Since(collectionTime).Hours()) / (24 * intervalDays)
		toleranceCounter -= ((curIntervalIdx - lastIntervalIdx) - 1)
		// fmt.Printf("collection time: %s, cur interval idx: %d, last interval idx: %d, tolerance counter: %d\n", collectionTime, curIntervalIdx, lastIntervalIdx, toleranceCounter)

		if toleranceCounter < 0 {
			return false
		}
		lastIntervalIdx = curIntervalIdx
	}
	return true
}

func isActiveWatching(bgmAPI *dao.BgmApiAccessor, uid string) bool {
	watching, err := bgmAPI.GetRecentCollections(uid, model.Watching, model.Anime, animeFilter, util.ActivityCheckDays)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return len(watching) >= util.MinWatchingCnt
}

func animeFilter(animeCol gjson.Result) bool {
	tags := animeCol.Get("subject").Get("tags").Array()
	for _, tag := range tags {
		if _, ok := tagsToReject[tag.Get("name").String()]; ok {
			// fmt.Printf("Rejecting collection with tag: %s\n", tag.Get("name").String())
			return false
		}
	}
	// only accept collection with rating
	rating := int(animeCol.Get("rate").Int())
	if rating == 0 {
		// fmt.Printf("Rejecting collection with rating: %d\n", rating)
		return false
	}
	collectionTotal := animeCol.Get("subject").Get("collection_total").Int()
	// assuming a subject with too few collections are not generally available
	// meaning not watching it does not necessarily mean people are not interested in the work
	return collectionTotal >= util.SubjectMinCollectionCnt
}
