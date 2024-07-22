package param

import (
	"flag"
	"fmt"
	"time"
)

type Params struct {
	Mode CrawlerMode
}

func GetParams() (params Params) {
	var modeStr string
	flag.StringVar(&modeStr, "mode", "", "mode: subject or user")
	flag.Parse()

	return Params{
		Mode: getMode(modeStr),
	}
}

func getMode(modeStr string) CrawlerMode {
	if mode, err := CrawlerModeFromString(modeStr); err == nil {
		fmt.Printf("Retrieving CrawlerMode from mode arg: %s", modeStr)
		return mode
	} else {
		fmt.Println("Retrieving CrawlerMode from month")
		var month = time.Now().Month()
		switch int(month) % 3 {
		case 0:
			fmt.Printf("Return SubjectMode for month: %d", month)
			return SubjectMode
		default:
			fmt.Printf("Return UserMode for month: %d", month)
			return UserMode
		}
	}
}
