package main

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	foxFlag            = kingpin.Flag("mf", "use mangafox as source").Bool()
	readerFlag         = kingpin.Flag("mr", "use mangaReader as source").Bool()
	vlm                = kingpin.Flag("vlm", "use when you want to download a volume(s)").Bool()
	update             = kingpin.Flag("update", "use to update the manga in your lib to the latest chapter").Bool()
	maxActiveDownloads = kingpin.Flag("n", "set max number of concurrent downloads").Int()
	manga              = kingpin.Arg("manga", "The name of the manga").String()
	args               = kingpin.Arg("arguments",
		"chapters (volumes if --vlm is set) to download. Example format: 1 3 5-7 67 10-14").Strings()
)

const (
	readerURL = "http://www.mangareader.net"
	foxURL    = "http://mangafox.me/"
)

func main() {
	kingpin.Parse()
	n := 1 //default num of maxActiveDownloads
	if *maxActiveDownloads != 0 {
		n = *maxActiveDownloads
	}

	if *foxFlag {
		n = 1 // so we don't hammer the mangafox site...we start getting errors if we set this any higher
		fmt.Println("Setting max active downloads to 1. You will get errors or missing chapter images if you set it any higher with mangafox as source.")
	}

	if len(*args) <= 0 && !*update {
		fmt.Println("See GoManga --help for usage tips")
		return
	}

	var source mangaSource

	switch {
	case *foxFlag:
		source = &foxManga{
			Args:      GetRange(args),
			MangaName: manga,
			sourceUrl: foxURL,
		}
		if *vlm { //if we're downloading volumes
			source.getVolumes(n)
		} else {
			source.getChapters(n)
		}
	case *readerFlag:
		source = &readerManga{
			Args:      GetRange(args),
			MangaName: manga,
			sourceUrl: readerURL,
		}
		source.getChapters(n)
	case *update:
		updateMangaLib()
	default:
		fmt.Println("Default source used: MangaReader.")
		source = &readerManga{
			Args:      GetRange(args),
			MangaName: manga,
			sourceUrl: readerURL,
		}
		source.getChapters(n)
	}

}
