package main

type mangaSource interface {
	search() (map[int]searchResult, error)
	getChapters(n int)
	getVolumes(n int)
}

type chapterSource interface {
	getChapter() error
}

type searchResult struct {
	manga, mangaID string
}
