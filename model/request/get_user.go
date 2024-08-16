package request

const (
	getGetUserUriPrefix = "/v0/users/"
)

type GetUserRequest struct {
	Uid string
}

func (request *GetUserRequest) ToUri() string {
	return getGetUserUriPrefix + request.Uid
}
