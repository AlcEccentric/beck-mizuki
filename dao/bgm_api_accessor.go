package dao

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	model "github.com/alceccentric/beck-crawler/model"
	req "github.com/alceccentric/beck-crawler/model/request"
	util "github.com/alceccentric/beck-crawler/util"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type BgmApiAccessor struct {
	httpClient *resty.Client
	randGen    *rand.Rand
}

func NewBgmApiAccessor() *BgmApiAccessor {
	return &BgmApiAccessor{
		httpClient: resty.New(),
		randGen:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (apiClient *BgmApiAccessor) GetSubjects(tags []string, types []model.SubjectType, airDateRange [2]time.Time, ratingRange [2]float32) ([]model.Subject, error) {
	offset := 0
	subjects := make([]model.Subject, 0)
	for {
		// fmt.Printf("Sending get subjects request with time range %s and offset %d\n", airDateRange, offset)
		respBody, resp, err := apiClient.post(&req.SearchSubjectPagedRequest{
			Tags:         tags,
			Types:        types,
			AirDateRange: airDateRange,
			RatingRange:  ratingRange,
			Offset:       offset,
			Limit:        util.PageLimit,
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
		if len(subjectResults) < util.PageLimit {
			break
		}
		offset += util.PageLimit
	}
	return subjects, nil
}

func (apiClient *BgmApiAccessor) GetUser(uid string) (model.User, error) {
	// fmt.Printf("Sending get user request\n")
	getUserResult, resp, getUserErr := apiClient.get(&req.GetUserRequest{
		Uid: uid,
	})
	// index 0 points to the latest collection
	latestCollectionTime, getLatestCollectionErr := apiClient.GetCollectionTime(uid, 0)

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

func (apiClient *BgmApiAccessor) GetCollections(uid string, ctype model.CollectionType, stype model.SubjectType, collectionAcceptor func(gjson.Result) bool) ([]model.Collection, error) {
	offset := 0
	collections := make([]model.Collection, 0)
	for {
		originalColCnt := len(collections)
		var err error
		collections, err = apiClient.addCollections(&req.GetPagedUserCollectionsRequest{
			Uid:            uid,
			CollectionType: ctype,
			SubjectType:    stype,
			Offset:         offset,
			Limit:          util.PageLimit,
		}, collectionAcceptor, collections)

		if err != nil {
			return nil, err
		}

		if len(collections) == originalColCnt {
			break
		}
		offset += util.PageLimit
	}
	return collections, nil
}

func (apiClient *BgmApiAccessor) GetRecentCollections(uid string,
	ctype model.CollectionType,
	stype model.SubjectType,
	collectionAcceptor func(gjson.Result) bool,
	recentWindowInDays int) ([]model.Collection, error) {
	offset := 0
	collections := make([]model.Collection, 0)
	for {
		originalColCnt := len(collections)
		var fetchErr error
		collections, fetchErr = apiClient.addCollections(&req.GetPagedUserCollectionsRequest{
			Uid:            uid,
			CollectionType: ctype,
			SubjectType:    stype,
			Offset:         offset,
			Limit:          util.PageLimit,
		}, collectionAcceptor, collections)
		if fetchErr != nil {
			return nil, fetchErr
		}

		if len(collections) == originalColCnt {
			break
		}

		oldestCollectionTime, parseErr := time.Parse(util.CollectionTimeFormat, collections[len(collections)-1].CollectedTime)
		if parseErr != nil {
			return nil, parseErr
		}

		if time.Since(oldestCollectionTime) >= time.Duration(recentWindowInDays)*24*time.Hour {
			break
		}
		offset += util.PageLimit
	}
	return collections, nil
}

func (apiClient *BgmApiAccessor) addCollections(getPagedCollectionReq *req.GetPagedUserCollectionsRequest,
	collectionAcceptor func(gjson.Result) bool,
	collections []model.Collection) ([]model.Collection, error) {

	// fmt.Printf("Sending get collections request\n")
	respBody, resp, err := apiClient.get(getPagedCollectionReq)
	if err != nil {
		return collections, err
	} else if exceedMaxCollectionCnt(resp, respBody) {
		return collections, nil
	} else if !resp.IsSuccess() {
		return collections, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}

	collectionResults := respBody.Get("data").Array()
	for _, collectionResult := range collectionResults {
		if collectionAcceptor(collectionResult) {
			collections = append(collections, model.Collection{
				UserID:         getPagedCollectionReq.Uid,
				SubjectType:    int(getPagedCollectionReq.SubjectType),
				SubjectID:      collectionResult.Get("subject_id").String(),
				CollectionType: int(getPagedCollectionReq.CollectionType),
				CollectedTime:  collectionResult.Get("updated_at").String(),
				Rating:         int(collectionResult.Get("rate").Int()),
			})
		}
	}
	return collections, nil
}

func (apiClient *BgmApiAccessor) GetCollectionCount(uid string, ctype model.CollectionType, stype model.SubjectType) (int, error) {
	request := &req.GetPagedUserCollectionsRequest{
		Uid:            uid,
		CollectionType: ctype,
		SubjectType:    stype,
		Limit:          1,
		Offset:         util.MaxWatchedAnimeCount,
	}

	// fmt.Printf("Sending get collection count request\n")
	respBody, resp, err := apiClient.get(request)

	if err != nil {
		return 0, err
	}

	if resp.IsSuccess() {
		return 0, fmt.Errorf("found a moron %s who said she/he watched more than %d animes", uid, util.MaxWatchedAnimeCount)
	} else if exceedMaxCollectionCnt(resp, respBody) {
		re := regexp.MustCompile(`less than or equal to (\d+)`)
		match := re.FindStringSubmatch(respBody.Get("description").String())
		if len(match) > 1 {
			if count, err := strconv.Atoi(match[1]); err == nil {
				return count, nil
			} else {
				return 0, err
			}
		} else {
			return 0, fmt.Errorf("not able to find collection count from message %s", respBody.Get("description").String())
		}
	} else {
		return 0, fmt.Errorf("request failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}
}

func (apiClient *BgmApiAccessor) GetCollectionTime(uid string, offset int) (time.Time, error) {
	getLatestCollectionRequest := &req.GetPagedUserCollectionsRequest{
		Uid:            uid,
		CollectionType: model.Watched,
		SubjectType:    model.Anime,
		Limit:          1,
		Offset:         offset,
	}

	// fmt.Printf("Sending get latest collection request\n")
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
	apiClient.randDelay()
	resp, err := apiClient.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "alceccentric/beck-crawler").
		Get(util.ApiDomain + request.ToUri())
	if err != nil {
		return gjson.Result{}, nil, err
	}

	return gjson.ParseBytes(resp.Body()), resp, nil
}

func (apiClient *BgmApiAccessor) post(request req.BgmPostRequest) (gjson.Result, *resty.Response, error) {
	apiClient.randDelay()
	resp, err := apiClient.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "alceccentric/beck-crawler").
		SetBody(request.ToBody()).
		Post(util.ApiDomain + request.ToUri())
	if err != nil {
		return gjson.Result{}, nil, err
	}

	return gjson.ParseBytes(resp.Body()), resp, nil
}

func (apiClient *BgmApiAccessor) randDelay() {
	time.Sleep(time.Duration((apiClient.randGen.Intn(500) + 500)) * time.Millisecond)
}

func exceedMaxCollectionCnt(resp *resty.Response, respBody gjson.Result) bool {
	return resp.StatusCode() == 400 && strings.Contains(respBody.Get("description").String(), "offset should be less than or equal to")
}
