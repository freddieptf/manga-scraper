package main

import (
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/fanfox"
	"github.com/freddieptf/manga-scraper/pkg/mangareader"
	"github.com/freddieptf/manga-scraper/pkg/mangastream"
	"os"
)

var (
	foxFlag            = flag.Bool("mf", false, "search mangafox for the manga")
	readerFlag         = flag.Bool("mr", false, "search mangaReader for the manga")
	mangastreamFlag    = flag.Bool("ms", false, "search mangastream for the manga")
	vlm                = flag.Bool("vlm", false, "use with -mf when you want to download a volume(s)")
	update             = flag.Bool("update", false, "use to update the manga in your local library to the latest chapter")
	archive            = flag.Bool("cbz", true, "save to cbz file format")
	maxActiveDownloads = flag.Int("n", 1, "max number of concurrent downloads")
	manga              = flag.String("manga", "", "the name of the manga")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if !*foxFlag && !*readerFlag && !*mangastreamFlag {
		fmt.Println("No source was provided.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	n := 1 //default num of maxActiveDownloads
	var source cli.MangaSource

	switch {
	case *foxFlag:
		source = &fanfox.FoxManga{}
	case *readerFlag:
		if *maxActiveDownloads != 0 {
			n = *maxActiveDownloads
		}
		source = &mangareader.ReaderManga{}
	case *mangastreamFlag:
		source = &mangastream.MangaStream{}
	default:
		flag.PrintDefaults()
		return
	}

	cli.Get(n, *manga, cli.GetChapterRangeFromArgs(&args), archive, source)

}
