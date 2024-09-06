package helper

import (
	"fmt"
	"time"

	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/model"
	"github.com/alceccentric/beck-crawler/util"
	"github.com/tidwall/gjson"
)

// Reject users with less than 400 raw collections
// Reject users whose oldest collection was made in the last 12 months
// For users with more than 400 but less than 800 collections, require them to have at least 1 collection every 15 days in the past half year
// For users with more than 800 but less than 1200 collections, require them to have at least 1 collection every 30 days in the past half year
// For users with more than 1200 collections, require them to have at least 1 collection every 45 days in the past half year
// For each user, up to 2 periods are allowed without a collection
// Reject users with less than 300 filtered collections

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
	// raw collection count filter
	rawCollectionCount, err := bgmAPI.GetCollectionCount(uid, model.Watched, model.Anime)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if rawCollectionCount < util.T1CollectionCnt {
		fmt.Println(fmt.Errorf("user: %s raw collection count was under %d", uid, util.T1CollectionCnt))
		return false, nil
	}

	// oldest collection filter
	earliestCollectionTime, err := bgmAPI.GetCollectionTime(uid, rawCollectionCount-1)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if time.Since(earliestCollectionTime) < time.Hour*24*util.MinOldestCollectionAgeInDays {
		fmt.Println(fmt.Errorf("user: %s earliest collection time was under %d days from today", uid, util.MinOldestCollectionAgeInDays))
		return false, nil
	}

	// leveled activity check
	colletionsInPastHalfYear, err := bgmAPI.GetRecentCollections(uid, model.Watched, model.Anime, animeFilter, util.ActivityCheckDays)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	active := true
	if rawCollectionCount < util.T2CollectionCnt {
		active = isActive(colletionsInPastHalfYear, util.T1IntervalDays)
	} else if rawCollectionCount < util.T3CollectionCnt {
		active = isActive(colletionsInPastHalfYear, util.T2IntervalDays)
	} else {
		active = isActive(colletionsInPastHalfYear, util.T3IntervalDays)
	}

	if !active {
		fmt.Printf("user: %s with %d raw collections is not active\n", uid, rawCollectionCount)
		return false, nil
	}

	// filtered collection count check
	filteredCollections, err := bgmAPI.GetCollections(uid, model.Watched, model.Anime, animeFilter)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if len(filteredCollections) < util.MinFilteredCollectionCnt {
		fmt.Printf("user: %s with %d filtered collections is under %d\n", uid, len(filteredCollections), util.MinFilteredCollectionCnt)
		return false, nil
	}

	return true, filteredCollections
}

func isActive(collections []model.Collection, intervalDays int) bool {
	lastIntervalIdx := -1
	toleranceCounter := util.InActiveIntervalTolerance
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
