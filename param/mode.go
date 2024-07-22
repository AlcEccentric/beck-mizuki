package param

import "fmt"

// CrawlerMode
type CrawlerMode int

const (
	SubjectMode CrawlerMode = iota
	UserMode
)

func CrawlerModeFromString(modeStr string) (mode CrawlerMode, err error) {
	switch modeStr {
	case "subject":
		return SubjectMode, nil
	case "user":
		return UserMode, nil
	default:
		return -1, fmt.Errorf("mode %s is not supported", modeStr)
	}
}
