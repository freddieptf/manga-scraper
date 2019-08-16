package scraper

type ChapterPage struct {
	Page int    `json:"page"`
	Url  string `json:"url"`
}

type Chapter struct {
	ID           string        `json:"id"`
	ChapterTitle string        `json:"title"`
	URL          string        `json:"url"`
	MangaName    string        `json:"manga"`
	VolumeTitle  string        `json:"volume,omitEmpty"`
	SourceName   string        `json:"-"`
	ChapterPages []ChapterPage `json:"pages,omitEmpty"`
}

type Manga struct {
	MangaName string `json:"name"`
	MangaID   string `json:"id"`
}
