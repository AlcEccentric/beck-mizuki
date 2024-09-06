package util

const (
	// user/collection filter parameters
	MinOldestCollectionAgeInDays = 365
	T1CollectionCnt              = 400
	T2CollectionCnt              = 800
	T3CollectionCnt              = 1200
	ActivityCheckDays            = 180
	T1IntervalDays               = 10
	T2IntervalDays               = 20
	T3IntervalDays               = 30
	InActiveIntervalTolerance    = 3
	MinFilteredCollectionCnt     = 300
	SubjectMinCollectionCnt      = 100
	MaxWatchedAnimeCount         = 3000

	// various data format
	SubjectDateFormat           = "2006-01-02"
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
	ColdStartIntervalInDays = 7
)
