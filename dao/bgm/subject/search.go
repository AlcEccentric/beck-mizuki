package subject

import (
	"fmt"
	"strconv"

	"github.com/alceccentric/beck-crawler/dao/bgm"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type SubjectFetcher struct {
	client *resty.Client
}

func NewSubjectFetcher() SubjectFetcher {
	return SubjectFetcher{
		client: resty.New(),
	}
}

func (fetcher *SubjectFetcher) GetSubjects(request SubjectSearchRequest, subjectCh chan<- Subject) {
	offset := 0
	for {
		subjects, err := fetcher.searchSubjects(request, offset)
		if err != nil {
			panic(err)
		}
		for _, subject := range subjects {
			subjectCh <- Subject{
				Id:        subject.Get("id").String(),
				Type:      SubjectType(subject.Get("type").Int()),
				Name:      subject.Get("name").String(),
				AvgRating: float32(subject.Get("score").Float()),
			}
		}
		if len(subjects) < bgm.PageSize {
			break
		}
		offset += bgm.PageSize
	}
}

func (fetcher *SubjectFetcher) searchSubjects(request SubjectSearchRequest, offset int) (subjects []gjson.Result, err error) {
	resp, err := fetcher.client.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(request.ToBody()).
		Post(bgm.ApiDomain + bgm.SearchSubjectsUri + "?limit=" + string(bgm.PageSize) + "&offset=" + string(offset))

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}

	return gjson.GetBytes(resp.Body(), "data").Array(), nil
}

func SubjectTypeToString(a []SubjectType) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(int(v))
	}
	return b
}
