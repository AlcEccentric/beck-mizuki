package request

import (
	"fmt"
	"strconv"

	model "github.com/AlcEccentric/beck-mizuki/model"
)

type GetPagedUserCollectionsRequest struct {
	Uid            string
	CollectionType model.CollectionType
	SubjectType    model.SubjectType
	Limit          int
	Offset         int
}

func (request *GetPagedUserCollectionsRequest) ToUri() string {
	return getGetUserCollectionsUriPrefix(request.Uid) +
		"?subject_type=" + strconv.Itoa(int(request.SubjectType)) +
		"&type=" + strconv.Itoa(int(request.CollectionType)) +
		"&limit=" + fmt.Sprintf("%d", request.Limit) +
		"&offset=" + fmt.Sprintf("%d", request.Offset)
}

func getGetUserCollectionsUriPrefix(uid string) string {
	return "/v0/users/" + uid + "/collections"
}
