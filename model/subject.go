package model

type SubjectType int

// Define constants using iota
const (
	_ SubjectType = iota
	Manga
	Anime
	_
	Game
)

type Subject struct {
	Id        string
	Type      SubjectType
	Name      string
	AvgRating float32
}

func (st SubjectType) String() string {
	switch st {
	case Manga:
		return "Manga"
	case Anime:
		return "Anime"
	case Game:
		return "Game"
	default:
		return "Unknown"
	}
}
