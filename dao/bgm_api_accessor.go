package dao

import (
	"fmt"
	"strings"
	"time"

	model "github.com/alceccentric/beck-crawler/model"
	req "github.com/alceccentric/beck-crawler/model/request"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const (
	ApiDomain = "https://api.bgm.tv"
	pageLimit = 50
)

type BgmApiAccessor struct {
	httpClient *resty.Client
}

func NewBgmApiAccessor() BgmApiAccessor {
	return BgmApiAccessor{
		httpClient: resty.New(),
	}
}

func (apiClient *BgmApiAccessor) GetSubjects(tags []string, types []model.SubjectType, airDateRange [2]time.Time, ratingRange [2]float32) ([]model.Subject, error) {
	offset := 0
	subjects := make([]model.Subject, 0)
	for {
		respBody, resp, err := apiClient.post(&req.SearchSubjectPagedRequest{
			Tags:         tags,
			Types:        types,
			AirDateRange: airDateRange,
			RatingRange:  ratingRange,
			Offset:       offset,
			Limit:        pageLimit,
		})

		if err != nil {
			return nil, err
		}
		if resp.IsError() {
			return nil, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
		}

		subjectResults := respBody.Get("data").Array()
		for _, subjectResult := range subjectResults {
			subjects = append(subjects, model.Subject{
				Id:        subjectResult.Get("id").String(),
				Type:      model.SubjectType(subjectResult.Get("type").Int()),
				Name:      subjectResult.Get("name").String(),
				AvgRating: float32(subjectResult.Get("score").Float()),
			})
		}
		if len(subjectResults) < pageLimit {
			break
		}
		offset += pageLimit
	}
	return subjects, nil
}

func (apiClient *BgmApiAccessor) GetUser(uid string) (model.User, error) {
	getUserResult, resp, getUserErr := apiClient.get(&req.GetUserRequest{
		Uid: uid,
	})
	latestCollectionTime, getLatestCollectionErr := apiClient.getLatestCollectionTime(uid)

	if getUserErr != nil {
		return model.User{}, getUserErr
	} else if getLatestCollectionErr != nil {
		return model.User{}, getLatestCollectionErr
	} else if resp.StatusCode() != 200 {
		return model.User{}, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	} else {
		return model.User{
			ID:             uid,
			Nickname:       getUserResult.Get("nickname").String(),
			AvatarURL:      getUserResult.Get("avatar").Get("large").String(),
			LastActiveTime: latestCollectionTime,
		}, nil
	}
}

func (apiClient *BgmApiAccessor) GetUserCollections(uid string, ctype model.CollectionType, stype model.SubjectType) ([]model.Collection, error) {
	offset := 0
	collections := make([]model.Collection, 0)
	for {
		respBody, resp, err := apiClient.get(&req.GetPagedUserCollectionsRequest{
			Uid:            uid,
			CollectionType: ctype,
			SubjectType:    stype,
			Offset:         offset,
			Limit:          pageLimit,
		})
		if err != nil {
			return nil, err
		}
		if resp.IsError() {
			return nil, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
		}

		collectionResults := respBody.Get("data").Array()
		for _, collectionResult := range collectionResults {
			collections = append(collections, model.Collection{
				UserID:         uid,
				SubjectType:    int(stype),
				SubjectID:      collectionResult.Get("subject_id").String(),
				CollectionType: int(ctype),
				CollectedTime:  collectionResult.Get("updated_at").String(),
				Rating:         int(collectionResult.Get("rate").Int()),
			})
		}
		if len(collectionResults) < pageLimit {
			break
		}
		offset += pageLimit
	}
	return collections, nil
}

func (apiClient *BgmApiAccessor) IsCollectionMoreThan(uid string, ctype model.CollectionType, stype model.SubjectType, count int) (bool, error) {
	request := &req.GetPagedUserCollectionsRequest{
		Uid:            uid,
		CollectionType: ctype,
		SubjectType:    stype,
		Limit:          1,
		Offset:         count,
	}

	respBody, resp, err := apiClient.get(request)

	if err != nil {
		return false, err
	}

	if resp.IsSuccess() {
		return true, nil
	} else if resp.StatusCode() == 400 && strings.Contains(respBody.Get("description").String(), "offset should be less than or equal to") {
		return false, nil
	} else {
		return false, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}
}

func (apiClient *BgmApiAccessor) getLatestCollectionTime(uid string) (time.Time, error) {
	getLatestCollectionRequest := &req.GetPagedUserCollectionsRequest{
		Uid:            uid,
		CollectionType: model.Watched,
		SubjectType:    model.AnimeType,
		Limit:          1,
		Offset:         0,
	}
	respBody, resp, err := apiClient.get(getLatestCollectionRequest)
	if err != nil {
		return time.Now(), err
	}

	if resp.IsError() {
		return time.Now(), fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}

	return respBody.Get("data").Array()[0].Get("updated_at").Time(), nil
}

func (apiClient *BgmApiAccessor) get(request req.BgmGetRequest) (gjson.Result, *resty.Response, error) {
	resp, err := apiClient.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "alceccentric/beck-crawler").
		Get(ApiDomain + request.ToUri())
	if err != nil {
		return gjson.Result{}, nil, err
	}

	return gjson.ParseBytes(resp.Body()), resp, nil
}

func (apiClient *BgmApiAccessor) post(request req.BgmPostRequest) (gjson.Result, *resty.Response, error) {
	resp, err := apiClient.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "alceccentric/beck-crawler").
		SetBody(request.ToBody()).
		Post(ApiDomain + request.ToUri())
	if err != nil {
		return gjson.Result{}, nil, err
	}

	return gjson.ParseBytes(resp.Body()), resp, nil
}
