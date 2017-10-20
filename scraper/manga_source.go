package mangascraper

type MangaContract interface {
	SetName(name string)
	SetID(id string)
}

type Manga struct {
	MangaName string
	MangaID   string
}

func (manga Manga) SetName(name string) {
	manga.MangaName = name
}

func (manga Manga) SetID(id string) {
	manga.MangaID = id
}

type MangaSource interface {
	Search() ([]Manga, error)
	ScrapeChapters(n int) *chan ScrapeResult
	ScrapeVolumes(n int)
}
