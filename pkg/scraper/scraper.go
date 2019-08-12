package scraper

type ChapterPage struct {
	Page int
	Url  string
}

type Chapter struct {
	MangaName    string
	ChapterTitle string
	ID           string
	URL          string
	VolumeTitle  string
	SourceName   string
	ChapterPages []ChapterPage
}

type Manga struct {
	MangaName string
	MangaID   string
}
