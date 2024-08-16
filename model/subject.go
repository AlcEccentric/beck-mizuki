package model

type SubjectType int

// Define constants using iota
const (
	_ SubjectType = iota
	MangaType
	AnimeType
	_
	GameType
)

type Subject struct {
	Id        string
	Type      SubjectType
	Name      string
	AvgRating float32
}
