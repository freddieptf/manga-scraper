package cli

import (
	"flag"
	"log"
	"os"
	"strconv"
	"sync"

	scraper "github.com/freddieptf/manga-scraper/pkg/scraper"
)

type CliConf struct {
	IsSourceFox           *bool
	IsSourceRdr           *bool
	Archive               *bool
	Vlms                  *bool
	MangaName             *string
	ChapterArgs           *[]string
	ParallelDownloadLimit *int
}

type MangaSource interface {
	Search(mangaName string) ([]scraper.Manga, error)
	GetChapter(mangaID, chapterID string) (scraper.Chapter, error)
}

// pass a config here with the already parsed flags...
func Run(conf CliConf) {

	n := 1 //default num of maxActiveDownloads
	var source MangaSource

	switch {
	case *conf.IsSourceFox:
		source = &scraper.FoxManga{}
	case *conf.IsSourceRdr:
		if *conf.ParallelDownloadLimit != 0 {
			n = *conf.ParallelDownloadLimit
		}
		source = &scraper.ReaderManga{}
	default:
		flag.PrintDefaults()
		return
	}

	get(n, *conf.MangaName, GetChapterRangeFromArgs(conf.ChapterArgs), conf.Archive, source)

}

func get(n int, mangaName string, chapters *[]int, archive *bool, source MangaSource) {

	results, err := source.Search(mangaName)
	if err != nil {
		log.Fatal(err)
	}

	result := GetMatchFromSearchResults(ReadWrite{os.Stdin, os.Stdout}, results)

	downloadJobChan := make(chan chapter, n)
	var wg sync.WaitGroup
	startDownloads(n, &wg, &downloadJobChan)

	for _, chID := range *chapters {
		ch, err := source.GetChapter(result.MangaID, strconv.Itoa(chID))
		if err != nil {
			log.Printf("err: owwie %s\n", err)
		} else {
			ch.MangaName = result.MangaName
			chWrap := chapter{Chapter: ch, archive: *archive}
			downloadJobChan <- chWrap
			wg.Add(1)
		}
	}

	wg.Wait()
}
