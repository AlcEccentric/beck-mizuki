package model

import "time"

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

const (
	crCollectionTable = "bgm_user_collection"
)

type Collection struct {
	UserID         string    `bson:"user_id" gorm:"column:user_id"`
	SubjectID      string    `bson:"subject_id" gorm:"column:subject_id"`
	SubjectType    int       `bson:"subject_type" gorm:"column:subject_type"`
	CollectionType int       `bson:"collection_type" gorm:"column:collection_type"`
	CollectedTime  time.Time `bson:"collected_time" gorm:"column:collected_time"`
	Rating         int       `bson:"rating,omitempty" gorm:"column:rating"`
}

// It's for gorm to identify the target table
func (Collection) TableName() string {
	return crCollectionTable
}
