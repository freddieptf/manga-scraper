package mangascraper

type ChapterPage struct {
	Page int
	Url  string
}

type Chapter struct {
	MangaName    string
	ChapterTitle string
	ChapterPages []ChapterPage
}

type sourceChapter interface {
	getChapter() (Chapter, error)
}
