package bgm

import (
	"fmt"
	"strconv"

	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const (
	ApiDomain         = "https://api.bgm.tv"
	SearchSubjectsUri = "/v0/search/subjects"
	PageSize          = 50
)

type BgmApiClient struct {
	httpClient *resty.Client
}

func NewBgmApiClient() BgmApiClient {
	return BgmApiClient{
		httpClient: resty.New(),
	}
}

func (apiClient *BgmApiClient) GetSubjects(request bgmModel.SubjectSearchRequest, subjectChan chan<- bgmModel.Subject, index int) {
	offset := 0
	for {
		subjects, err := apiClient.getSubjects(request, offset)
		if err != nil {
			panic(err)
		}
		for _, subject := range subjects {
			fmt.Printf("Producer: %d, put item: %d\n", index, subject.Get("id").Int())
			subjectChan <- bgmModel.Subject{
				Id:        subject.Get("id").String(),
				Type:      bgmModel.SubjectType(subject.Get("type").Int()),
				Name:      subject.Get("name").String(),
				AvgRating: float32(subject.Get("score").Float()),
			}
		}
		if len(subjects) < PageSize {
			break
		}
		offset += PageSize
	}
}

func (bgmClient *BgmApiClient) getSubjects(request bgmModel.SubjectSearchRequest, offset int) (subjects []gjson.Result, err error) {
	resp, err := bgmClient.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "alceccentric/beck-crawler").
		SetBody(request.ToBody()).
		Post(ApiDomain + SearchSubjectsUri + "?limit=" + strconv.Itoa(PageSize) + "&offset=" + strconv.Itoa(offset))

	fmt.Printf("Sending request\n")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}

	return gjson.GetBytes(resp.Body(), "data").Array(), nil
}
