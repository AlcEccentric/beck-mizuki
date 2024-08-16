package request

type BgmGetRequest interface {
	ToUri() string
}

type BgmPostRequest interface {
	ToBody() string
	ToUri() string
}
