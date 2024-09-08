package util

const (
	// user/collection filter parameters
	MinOldestWatchedAgeInDays   = 365
	T1WatchedCnt                = 400
	T2WatchedCnt                = 800
	T3WatchedCnt                = 1200
	MinWatchingCnt              = 10
	ActivityCheckDays           = 90 // Better to be the same as CollectionCheckDays
	T1IntervalDays              = 10
	T2IntervalDays              = 20
	T3IntervalDays              = 30
	NonWatchedIntervalTolerance = 3
	MinFilteredWatchedCnt       = 300
	SubjectMinCollectionCnt     = 100
	MaxWatchedAnimeCount        = 3000

	// various data format
	SubjectDateFormat           = "2006-01-02"
	LaunchDateFormat            = "2006-01-02"
	CollectionTimeFormat        = "2006-01-02T15:04:05-07:00"
	WebsiteCollectionTimeFormat = "2006-1-2 15:04"

	// API parameters
	PageLimit           = 50
	ApiDomain           = "https://api.bgm.tv"
	GetGetUserUriPrefix = "/v0/users/"

	// Scraper parameters
	SubjectCollectionUrlFormat = "https://bangumi.tv/subject/%s/collections?page=%d"

	// subject filter parameters
	// TODO: make these configurable
	EarliestSubjectDate = "2022-01-01"
	LatestSubjectDate   = "2022-01-10"

	// Orchestration parameters
	ColdStartIntervalInDays     = 90
	RegularUpdateIntervalInDays = 30
)
