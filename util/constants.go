package util

const (
	// user/collection filter parameters
	MinOldestWatchedAgeInDays   = 365
	T1WatchedCnt                = 400
	T2WatchedCnt                = 800
	T3WatchedCnt                = 1200
	MinWatchingCnt              = 10
	ActivityCheckDays           = 90 // Should be at least twice as long as RegularUpdateIntervalInDays
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
	CollecttedTimeFormat        = "2006-01-02T15:04:05-07:00"
	WebsiteCollectionTimeFormat = "2006-1-2 15:04"

	// API parameters
	PageLimit                  = 50
	ApiDomain                  = "https://api.bgm.tv"
	GetGetUserUriPrefix        = "/v0/users/"
	APICallBaseDelayInMs       = 500
	APICallAdditionalDelayInMs = 500

	// Scraper parameters
	SubjectCollectionUrlFormat = "https://bangumi.tv/subject/%s/collections?page=%d"
	ScraperBaseDelayInS        = 2
	ScraperAdditionalDelayInS  = 2

	// Orchestration parameters
	// Cold start
	ColdStartIntervalInDays                  = 120
	NumOfSubjectRetrievers                   = 50
	NumOfUserIdRetrievers                    = 1 // could be more than 1 but should be cautious as it will incur high pressure on the target website
	NumOfUserIdMergers                       = 1 // must be one as the ids will be merged into a map and map is not thread safe
	UserIdRetrieverCoolDownSecondsPerSubject = 3
	// Regular update
	RegularUpdateIntervalInDays = 30
	NumOfUserIDReaders          = 5
	NumOfUserUpdaters           = 5
	NumOfUserCleaners           = 5
)
