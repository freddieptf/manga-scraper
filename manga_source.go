package main

type mangaSource interface {
	search() (map[int]searchResult, error)
	getChapters(n int)
	getVolumes(n int)
}

type chapterDownload struct {
	sourceUrl  string
	mangaId    string
	chapterUrl string
	manga      string //name of the manga
	chapter    string
	volume     string //what volume the chapter belongs to, optional really
}

type searchResult struct {
	manga, mangaID string
}
