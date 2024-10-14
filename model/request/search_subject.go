package request

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	model "github.com/AlcEccentric/beck-mizuki/model"
	util "github.com/AlcEccentric/beck-mizuki/util"
)

type SearchSubjectPagedRequest struct {
	Tags         []string
	Types        []model.SubjectType
	AirDateRange [2]time.Time // [begin, end)
	RatingRange  [2]float32
	Limit        int
	Offset       int
}

func (request *SearchSubjectPagedRequest) ToBody() string {

	startDate := request.AirDateRange[0].Format(util.SubjectDateFormat)
	endDate := request.AirDateRange[1].Format(util.SubjectDateFormat)

	return `{
		"keyword": "",
		"sort": "",
		"filter": {
		  "type": [` + strings.Join(subjectTypesToString(request.Types), ",") + `],
		  "tag": ["` + strings.Join(request.Tags, ",") + `"],
		  "air_date": [
			">=` + startDate + `",
			"<` + endDate + `"
		  ],
		  "rating": [
			">=` + fmt.Sprintf("%f", request.RatingRange[0]) + `",
			"<=` + fmt.Sprintf("%f", request.RatingRange[1]) + `"
		  ],
		  "rank": []
		}
	  }`
}

func (request *SearchSubjectPagedRequest) ToUri() string {
	return getSearchSubjectUriPrefix() + "?limit=" + fmt.Sprintf("%d", request.Limit) +
		"&offset=" + fmt.Sprintf("%d", request.Offset)
}

func getSearchSubjectUriPrefix() string {
	return "/v0/search/subjects"
}

func subjectTypesToString(a []model.SubjectType) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(int(v))
	}
	return b
}
