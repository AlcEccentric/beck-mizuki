package request

import "github.com/alceccentric/beck-crawler/util"

const ()

type GetUserRequest struct {
	Uid string
}

func (request *GetUserRequest) ToUri() string {
	return util.GetGetUserUriPrefix + request.Uid
}
