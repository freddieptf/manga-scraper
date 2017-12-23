package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	scraper "github.com/freddieptf/manga-scraper/pkg/scraper"
)

var (
	foxFlag            = flag.Bool("mf", false, "search mangafox for the manga")
	readerFlag         = flag.Bool("mr", false, "search mangaReader for the manga")
	vlm                = flag.Bool("vlm", false, "use with -mf when you want to download a volume(s)")
	update             = flag.Bool("update", false, "use to update the manga in your local library to the latest chapter")
	maxActiveDownloads = flag.Int("n", 1, "max number of concurrent downloads")
	manga              = flag.String("manga", "", "the name of the manga")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if !*foxFlag && !*readerFlag {
		fmt.Println("No source was provided. See -h for usage. Using mangareader as default instead...")
		*readerFlag = true
	}

	n := 1 //default num of maxActiveDownloads
	var source scraper.MangaSource

	switch {
	case *foxFlag:
		fmt.Println("Setting max active downloads to 1. " +
			"You will get errors or missing chapter images if you set it any higher with mangafox as source.")
		source = &scraper.FoxManga{}
		source.SetManga(scraper.Manga{MangaName: *manga})
		source.SetArgs(getChapterRangeFromArgs(&args))
		if *vlm { //if we're downloading volumes
			getVolumes(n, source)
		} else {
			getChapters(n, source)

		}
	case *readerFlag:
		if *maxActiveDownloads != 0 {
			n = *maxActiveDownloads
		}
		source = &scraper.ReaderManga{}
		source.SetManga(scraper.Manga{MangaName: *manga})
		source.SetArgs(getChapterRangeFromArgs(&args))
		getChapters(n, source)
	default:
		flag.PrintDefaults()
		return
	}

}

func getChapters(n int, source scraper.MangaSource) {
	results, err := source.Search()
	if err != nil {
		log.Fatal(err)
	}
	result := getMatchFromSearchResults(readWrite{os.Stdin, os.Stdout}, results)
	source.SetManga(result)

	resultsChan := source.ScrapeChapters(n)
	startDownloads(n, len(*source.GetArgs()), resultsChan)
}

func getVolumes(n int, source scraper.MangaSource) {
	results, err := source.Search()
	if err != nil {
		log.Fatal(err)
	}
	result := getMatchFromSearchResults(readWrite{os.Stdin, os.Stdout}, results)
	source.SetManga(result)

	count, resultsChan := source.ScrapeVolumes(n)
	if count.Err != nil {
		log.Fatalf("Encountered an error while trying to scrape volumes : err %v\n", count.Err)
	}
	startDownloads(n, count.ChapterCount, resultsChan)
}
