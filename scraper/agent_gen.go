package scraper

import (
	"math/rand"
	"time"
)

var userAgentList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
}

type UserAgentGenerator struct {
	randGen *rand.Rand
}

func NewUserAgentGenerator() *UserAgentGenerator {
	return &UserAgentGenerator{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (agentGen *UserAgentGenerator) RandomUserAgent() string {
	return userAgentList[agentGen.randGen.Intn(len(userAgentList))]
}
