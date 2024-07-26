package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SubjectSearchRequest struct {
	Tag          []string
	Type         []SubjectType
	AirDateRange [2]time.Time
	RatingRange  [2]float32
}

const (
	DateFormat = "2006-01-02"
)

func (request *SubjectSearchRequest) ToBody() string {

	startDate := request.AirDateRange[0].Format(DateFormat)
	endDate := request.AirDateRange[1].Format(DateFormat)

	return `{
		"keyword": "",
		"sort": "",
		"filter": {
		  "type": [` + strings.Join(subjectTypeToString(request.Type), ",") + `],
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

func subjectTypeToString(a []SubjectType) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(int(v))
	}
	return b
}
