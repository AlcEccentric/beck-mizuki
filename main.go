package main

import (
	scraper "github.com/alceccentric/beck-crawler/scraper"
)

// const (
// 	numOfSubjectProducers    = 3
// 	numOfUserProducers       = 3
// 	numOfCollectionProducers = 3
// )

func main() {
	// bgmClient := bgm.NewBgmApiClient()
	// subjectOrch := orch.NewSubjectOrchestrator(&bgmClient, numOfCollectionProducers, numOfUserProducers, numOfSubjectProducers)
	// subjectOrch.Run()
	subjectUserCrawler := scraper.NewSubjectUserScraper()
	subjectUserCrawler.Crawl("242745")
}
