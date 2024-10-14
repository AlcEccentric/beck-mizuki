package model

import (
	"time"

	jetmodel "github.com/AlcEccentric/beck-mizuki/model/gen/beck-konomi/public/model"
)

type CollectionType int

const (
	_ CollectionType = iota
	ToWatch
	Watched
	Watching
	Postponed
	Discarded
)

func (ct CollectionType) String() string {
	switch ct {
	case ToWatch:
		return "ToWatch"
	case Watched:
		return "Watched"
	case Watching:
		return "Watching"
	case Postponed:
		return "Postponed"
	case Discarded:
		return "Discarded"
	default:
		return "Unknown"
	}
}

type Collection struct {
	UserID         string    `bson:"user_id" gorm:"column:user_id"`
	SubjectID      string    `bson:"subject_id" gorm:"column:subject_id"`
	SubjectType    int64     `bson:"subject_type" gorm:"column:subject_type"`
	CollectionType int64     `bson:"collection_type" gorm:"column:collection_type"`
	CollectedTime  time.Time `bson:"collected_time" gorm:"column:collected_time"`
	Rating         int64     `bson:"rating,omitempty" gorm:"column:rating"`
}

func (c *Collection) ToBgmUserCollection() jetmodel.BgmUserCollection {
	return jetmodel.BgmUserCollection{
		UserID:         c.UserID,
		SubjectID:      c.SubjectID,
		SubjectType:    &c.SubjectType,
		CollectionType: &c.CollectionType,
		CollectedTime:  &c.CollectedTime,
		Rating:         &c.Rating,
	}
}

func ToBgmUserCollections(collections []Collection) []jetmodel.BgmUserCollection {
	bgmUserCollections := make([]jetmodel.BgmUserCollection, 0, len(collections))
	for _, collection := range collections {
		bgmUserCollections = append(bgmUserCollections, collection.ToBgmUserCollection())
	}
	return bgmUserCollections
}
