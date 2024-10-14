package request

import "github.com/AlcEccentric/beck-mizuki/util"

const ()

type GetUserRequest struct {
	Uid string
}

func (request *GetUserRequest) ToUri() string {
	return util.GetGetUserUriPrefix + request.Uid
}
