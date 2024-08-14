package model

import "time"

type UserCollection struct {
	UserID         string    `bson:"user_id"`
	SubjectID      string    `bson:"subject_id"`
	SubjectType    int       `bson:"subject_type"`
	Rating         float32   `bson:"rating,omitempty"`
	CollectionType int       `bson:"collection_type"`
	CollectedTime  time.Time `bson:"collected_time"`
}
