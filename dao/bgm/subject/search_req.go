package subject

import (
	"fmt"
	"strings"
	"time"

	"github.com/alceccentric/beck-crawler/dao/bgm"
)

type SubjectSearchRequest struct {
	Tag          []string
	Type         []SubjectType
	AirDateRange [2]time.Time
	RatingRange  [2]float32
}

func (request *SubjectSearchRequest) ToBody() string {

	startDate := request.AirDateRange[0].Format(bgm.DateFormat)
	endDate := request.AirDateRange[1].Format(bgm.DateFormat)

	return `{
		"keyword": "",
		"sort": "",
		"filter": {
		  "type": [` + strings.Join(SubjectTypeToString(request.Type), ",") + `],
		  "tag": ["` + strings.Join(request.Tag, ",") + `"],
		  "air_date": [
			">=` + startDate + `",
			"<=` + endDate + `"
		  ],
		  "rating": [
			">=` + fmt.Sprintf("%f", request.RatingRange[0]) + `",
			"<=` + fmt.Sprintf("%f", request.RatingRange[1]) + `"
		  ],
		  "rank": []
		}
	  }`
}
