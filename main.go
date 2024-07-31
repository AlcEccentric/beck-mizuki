package main

import (
	"github.com/alceccentric/beck-crawler/dao/bgm"
	"github.com/alceccentric/beck-crawler/orch"
)

const (
	numOfSubjectProducers    = 3
	numOfUserProducers       = 3
	numOfCollectionProducers = 3
)

func main() {
	bgmClient := bgm.NewBgmApiClient()
	subjectOrch := orch.NewSubjectOrchestrator(&bgmClient, numOfCollectionProducers, numOfUserProducers, numOfSubjectProducers)
	subjectOrch.Run()
}
