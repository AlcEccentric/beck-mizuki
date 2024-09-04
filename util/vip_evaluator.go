package util

import (
	"fmt"
	"time"

	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/model"
	"github.com/tidwall/gjson"
)

// Reject users with less than 400 raw collections
// Reject users whose oldest collection was made in the last 12 months
// For users with more than 400 but less than 800 collections, require them to have at least 1 collection every 10 days in the past half year
// For users with more than 800 but less than 1200 collections, require them to have at least 1 collection every 20 days in the past half year
// For users with more than 1200 collections, require them to have at least 1 collection every 30 days in the past half year
// For each user, up to 2 periods are allowed without a collection
// Reject users with less than 300 filtered collections

const (
	minOldestCollectionAgeInDays = 365
	t1CollectionCnt              = 400
	t2CollectionCnt              = 800
	t3CollectionCnt              = 1200
	activityCheckDays            = 180
	t1IntervalDays               = 10
	t2IntervalDays               = 20
	t3IntervalDays               = 30
	inActiveIntervalTolerance    = 2
	minFilteredCollectionCnt     = 300
	subjectMinCollectionCnt      = 100
	collectionTimeFormat         = "2006-01-02"
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

func IsVip(uid string, bgmAPI *dao.BgmApiAccessor) (bool, []model.Collection) {
	// raw collection count filter
	rawCollectionCount, err := bgmAPI.GetCollectionCount(uid, model.Watched, model.Anime)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if rawCollectionCount < t1CollectionCnt {
		fmt.Println(fmt.Errorf("user: %s raw collection coutn was under %d", uid, t1CollectionCnt))
		return false, nil
	}

	// oldest collection filter
	earliestCollectionTime, err := bgmAPI.GetCollectionTime(uid, rawCollectionCount-1)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if time.Since(earliestCollectionTime) < time.Hour*24*minOldestCollectionAgeInDays {
		fmt.Println(fmt.Errorf("user: %s earliest collection time was under %d days from today", uid, minOldestCollectionAgeInDays))
		return false, nil
	}

	// leveled activity check
	colletionsInPastHalfYear, err := bgmAPI.GetRecentCollections(uid, model.Watched, model.Anime, animeFilter, activityCheckDays)

	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	active := true
	if rawCollectionCount < t2CollectionCnt {
		active = isActive(colletionsInPastHalfYear, t1IntervalDays)
	} else if rawCollectionCount < t3CollectionCnt {
		active = isActive(colletionsInPastHalfYear, t2IntervalDays)
	} else {
		active = isActive(colletionsInPastHalfYear, t3IntervalDays)
	}

	if !active {
		return false, nil
	}

	// filtered collection count check
	filteredCollections, err := bgmAPI.GetCollections(uid, model.Watched, model.Anime, animeFilter)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	if len(filteredCollections) < minFilteredCollectionCnt {
		return false, nil
	}

	return true, filteredCollections
}

func isActive(collections []model.Collection, intervalDays int) bool {
	lastIntervalIdx := -1
	toleranceCounter := inActiveIntervalTolerance
	for i := 0; i < len(collections); i++ {
		collectionTime, err := time.Parse(collectionTimeFormat, collections[i].CollectedTime)
		if err != nil {
			panic(fmt.Errorf("failed to parse collection time: %s (%w)", collections[i].CollectedTime, err))
		}
		curIntervalIdx := int(time.Since(collectionTime).Hours()) / (24 * intervalDays)
		toleranceCounter -= ((curIntervalIdx - lastIntervalIdx) - 1)

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
			return false
		}
	}
	collectionTotal := animeCol.Get("subject").Get("collection_total").Int()
	// assuming a subject with too few collections are not generally available
	// meaning not watching it does not necessarily mean people are not interested in the work
	return collectionTotal >= subjectMinCollectionCnt
}
