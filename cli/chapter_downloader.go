package cli

import (
	"log"
	"sync"
	"time"

	scraper "github.com/freddieptf/manga-scraper/scraper"
)

var downloadJobChan chan scraper.Chapter

type chapterDownloader struct {
	id   int
	quit chan bool
	wg   *sync.WaitGroup
}

func getChapters(n int, source scraper.MangaSource) {
	results, err := source.Search()
	if err != nil {
		log.Fatal(err)
	}
	result := getMatchFromSearchResults(results)
	source.SetManga(result)

	resultsChan := source.ScrapeChapters(n)
	startDownloads(n, len(*source.GetArgs()), resultsChan)
}

func newChapterDownloader(id int, wg *sync.WaitGroup) *chapterDownloader {
	return &chapterDownloader{
		id:   id,
		wg:   wg,
		quit: make(chan bool),
	}
}

// starts listening for jobs posted on the downloadJobChan channel
func (chDown *chapterDownloader) listen() {
	go func() {
		for {
			select {
			case chapter := <-downloadJobChan:
				log.Printf("downloader %d: Start new Download %v\n", chDown.id, chapter.ChapterTitle)
				time.Sleep(2 * time.Second)
				log.Printf("downloader %d: Download done %v\n", chDown.id, chapter.ChapterTitle)
				chDown.wg.Done()
			case <-chDown.quit:
				return
			}
		}
	}()
}

// params: 	n - the max active parallel downloads
//			total - the total no of downloads
//			resultsChan - the channel that provides the scrapeResult:{ either a chapter or an error }
func startDownloads(n, total int, resultsChan *chan scraper.ScrapeResult) {
	downloadJobChan = make(chan scraper.Chapter, n)
	var wg sync.WaitGroup

	// init our download workers right here boy
	for i := 0; i < n; i++ {
		chDownloader := newChapterDownloader(i+1, &wg)
		chDownloader.listen()
	}

	count := 0
	for {
		select {
		case scrapeResult := <-*resultsChan:
			count++
			wg.Add(1)
			if scrapeResult.Err != nil {
				log.Printf("Damn err, %v\n", scrapeResult.Err)
				wg.Done()
			} else {
				downloadJobChan <- scrapeResult.Chapter
				if count == total { // if the no of results we've got == the total no of downloads
					wg.Wait() // then we wait for the downloads to finish
					return    // and exit
				}
			}
		}
	}
}
