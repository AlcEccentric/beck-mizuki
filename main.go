package main

import (
	"fmt"
	"sync"
	"time"

	bgm "github.com/alceccentric/beck-crawler/dao/bgm"
	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
)

const (
	numOfSubjectProducers    = 3
	numOfUserProducers       = 3
	numOfCollectionProducers = 3
)

func main() {
	bgmClient := bgm.NewBgmApiClient()
	startDate := time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)

	c := make(chan bgmModel.Subject, numOfSubjectProducers*10)

	var pWg sync.WaitGroup
	daysPerProduer := int(endDate.Sub(startDate).Hours()/24) / numOfSubjectProducers
	for i := 0; i < numOfSubjectProducers; i++ {

		pWg.Add(1)
		index := i // capture loop variable
		go func(index int) {
			defer pWg.Done()

			curStartDate := startDate.Add(time.Duration(index*daysPerProduer) * 24 * time.Hour)
			curEndDate := curStartDate.Add(time.Duration(daysPerProduer) * 24 * time.Hour)
			fmt.Printf("Producer: %d, start date: %s, end date: %s\n", index, curStartDate, curEndDate)

			if curEndDate.After(endDate) {
				curEndDate = endDate
			}
			bgmClient.GetSubjects(bgmModel.SubjectSearchRequest{
				Tag:          []string{"日本动画"},
				Type:         []bgmModel.SubjectType{bgmModel.Anime},
				AirDateRange: [2]time.Time{curStartDate, curEndDate},
				RatingRange:  [2]float32{0, 10},
			}, c, index)
		}(index)
	}
	var cWg sync.WaitGroup
	for i := 0; i < numOfUserProducers; i++ {
		cWg.Add(1)
		go func() {
			defer cWg.Done()
			for subject := range c {
				fmt.Printf("Subject id: %s, name: %s\n", subject.Id, subject.Name)
			}
		}()
	}

	// Wait for all goroutines to finish
	pWg.Wait()
	close(c)
	cWg.Wait()
}
