package model

type CollectionType int

// Define constants using iota
const (
	_ CollectionType = iota
	ToWatch
	Watched
	Watching
	Postponed
	Discarded
)

type Collection struct {
	UserID         string `bson:"user_id"`
	SubjectID      string `bson:"subject_id"`
	SubjectType    int    `bson:"subject_type"`
	CollectionType int    `bson:"collection_type"`
	CollectedTime  string `bson:"collected_time"`
	Rating         int    `bson:"rating,omitempty"`
}
