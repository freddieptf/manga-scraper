package main

import (
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/mangareader"
	"os"
)

var (
	readerFlag         = flag.Bool("mr", false, "search mangaReader for the manga")
	update             = flag.Bool("update", false, "use to update the manga in your local library to the latest chapter")
	archive            = flag.Bool("cbz", true, "save to cbz file format")
	maxActiveDownloads = flag.Int("n", 1, "max number of concurrent downloads")
	manga              = flag.String("manga", "", "the name of the manga")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if !*readerFlag {
		fmt.Println("No source was provided.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	n := 1 //default num of maxActiveDownloads
	var source cli.MangaSource

	switch {
	case *readerFlag:
		if *maxActiveDownloads != 0 {
			n = *maxActiveDownloads
		}
		source = &mangareader.ReaderManga{}
	// TODO: add mangadex here...
	default:
		flag.PrintDefaults()
		return
	}

	cli.Get(n, *manga, cli.GetChapterRangeFromArgs(&args), archive, source)

}
