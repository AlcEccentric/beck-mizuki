package subject

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
