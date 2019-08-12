package main

import (
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/mangastream"
	"os"
)

var (
	archive            = flag.Bool("cbz", true, "save to cbz file format")
	maxActiveDownloads = flag.Int("n", 1, "max number of concurrent downloads")
	manga              = flag.String("manga", "", "the name of the manga")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if *manga == "" {
		fmt.Println("no manga name passed")
		flag.PrintDefaults()
		os.Exit(1)
	}
	source := &mangastream.MangaStream{}
	cli.Get(*maxActiveDownloads, *manga, cli.GetChapterRangeFromArgs(&args), archive, source)
}
