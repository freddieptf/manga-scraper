package mangascraper

import (
	"log"
)

type scrapeJob struct {
	chapter sourceChapter
}

type scraper struct {
	scrapeResultsChan *chan ScrapeResult
	quitChan          chan bool
}

type ScrapeResult struct {
	Chapter Chapter
	Err     error
}

var scrapeJobChan chan *scrapeJob

func startScraping(maxScrapers int, workQueue chan *scrapeJob, resultChan *chan ScrapeResult) {

	scrapeJobChan = make(chan *scrapeJob, maxScrapers)

	for i := 0; i < maxScrapers; i++ {
		scraper := newScraper(resultChan)
		scraper.starto()
	}

	go func() {
		for job := range workQueue {
			go func(job *scrapeJob) {
				scrapeJobChan <- job
			}(job)
		}
	}()
}

func newScraper(resultChan *chan ScrapeResult) *scraper {
	return &scraper{
		scrapeResultsChan: resultChan,
		quitChan:          make(chan bool),
	}
}

// start listening for work
func (scraper *scraper) starto() {
	go func() {
		for {
			select {
			case job := <-scrapeJobChan:
				ch, err := job.chapter.getChapter()
				*scraper.scrapeResultsChan <- ScrapeResult{Chapter: ch, Err: err}
			case <-scraper.quitChan:
				log.Printf("scraper received quit signal")
				// return, stop listening for work
				return
			}
		}
	}()
}

// stop listening for work and stop when whatever is running finishes
func (scraper *scraper) korosu() {
	go func() {
		scraper.quitChan <- true
	}()
}
