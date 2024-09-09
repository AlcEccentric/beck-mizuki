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
	"github.com/rs/zerolog/log"
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
		log.Debug().Msgf("Sending get subjects request with time range %s [offset %d]", airDateRange, offset)
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
			return nil, fmt.Errorf("SearchSubjectPagedRequest failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
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
	log.Debug().Msgf("Sending get user request with uid %s", uid)
	getUserResult, resp, getUserErr := apiClient.get(&req.GetUserRequest{
		Uid: uid,
	})

	latestCollectionTime, getLatestCollectionErr := apiClient.GetCollectionTime(uid, 0, model.Watched, model.Anime)

	if getUserErr != nil {
		return model.User{}, getUserErr
	} else if getLatestCollectionErr != nil {
		return model.User{}, getLatestCollectionErr
	} else if resp.StatusCode() != 200 {
		return model.User{}, fmt.Errorf("GetUserRequest failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
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
	log.Debug().Msgf("Sending get collection request with uid %s, ctype %s, stype %s", uid, ctype.String(), stype.String())
	for {
		originalColCnt := len(collections)
		var err error
		log.Debug().Msgf("Sending get collection request with uid %s, ctype %s, stype %s [offset %d]", uid, ctype.String(), stype.String(), offset)
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
	log.Debug().Msgf("Sending get recent collection request with uid %s, ctype %s, stype %s, recentWindowInDays %d", uid, ctype.String(), stype.String(), recentWindowInDays)
	for {
		originalColCnt := len(collections)
		var fetchErr error
		log.Debug().Msgf("Sending get collection request with uid %s, ctype %s, stype %s, recentWindowInDays %d [offset %d]", uid, ctype.String(), stype.String(), recentWindowInDays, offset)
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

		if time.Since(collections[len(collections)-1].CollectedTime) >= time.Duration(recentWindowInDays)*24*time.Hour {
			break
		}
		offset += util.PageLimit
	}
	return collections, nil
}

func (apiClient *BgmApiAccessor) addCollections(getPagedCollectionReq *req.GetPagedUserCollectionsRequest,
	collectionAcceptor func(gjson.Result) bool,
	collections []model.Collection) ([]model.Collection, error) {
	respBody, resp, err := apiClient.get(getPagedCollectionReq)
	if err != nil {
		return collections, err
	} else if exceedMaxCollectionCnt(resp, respBody) {
		return collections, nil
	} else if !resp.IsSuccess() {
		return collections, fmt.Errorf("GetPagedUserCollectionsRequest failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}

	collectionResults := respBody.Get("data").Array()
	for _, collectionResult := range collectionResults {
		if collectionAcceptor(collectionResult) {

			collectedTime, err := time.Parse(util.CollecttedTimeFormat, collectionResult.Get("updated_at").String())

			if err != nil {
				log.Error().Err(err).Msgf("Failed to parse collection time %s for user %s subject id %s", collectionResult.Get("updated_at").String(),
					getPagedCollectionReq.Uid, collectionResult.Get("subject_id").String())
				continue
			}

			collections = append(collections, model.Collection{
				UserID:         getPagedCollectionReq.Uid,
				SubjectType:    int(getPagedCollectionReq.SubjectType),
				SubjectID:      collectionResult.Get("subject_id").String(),
				CollectionType: int(getPagedCollectionReq.CollectionType),
				CollectedTime:  collectedTime,
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

	log.Debug().Msgf("Sending get collection count request with uid %s, ctype %s, stype %s", uid, ctype.String(), stype.String())
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
		return 0, fmt.Errorf("GetPagedUserCollectionsRequest failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
	}
}

func (apiClient *BgmApiAccessor) GetCollectionTime(uid string, offset int, ctype model.CollectionType, stype model.SubjectType) (time.Time, error) {
	getLatestCollectionRequest := &req.GetPagedUserCollectionsRequest{
		Uid:            uid,
		CollectionType: ctype,
		SubjectType:    stype,
		Limit:          1,
		Offset:         offset,
	}

	log.Debug().Msgf("Sending get collection time request with uid %s, offset %d", uid, offset)
	respBody, resp, err := apiClient.get(getLatestCollectionRequest)
	if err != nil {
		return time.Now(), err
	}

	if resp.IsError() {
		return time.Now(), fmt.Errorf("GetPagedUserCollectionsRequest failed with status: %s and code: %d", resp.Status(), resp.StatusCode())
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
	time.Sleep(time.Duration((apiClient.randGen.Intn(util.APICallAdditionalDelayInMs) + util.APICallBaseDelayInMs)) * time.Millisecond)
}

func exceedMaxCollectionCnt(resp *resty.Response, respBody gjson.Result) bool {
	return resp.StatusCode() == 400 && strings.Contains(respBody.Get("description").String(), "offset should be less than or equal to")
}
