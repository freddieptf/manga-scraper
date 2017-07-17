package main

import (
	s "GoManga/msources"
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

func main() {
	kingpin.Parse()
	n := 5 //default num of maxActiveDownloads
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

	download := s.MangaDownload{
		Args:      s.GetRange(args),
		MangaName: manga,
	}

	switch {
	case *foxFlag:
		if *vlm { //if we're downloading volumes
			download.GetVolumeFromFox(n)
		} else {
			download.GetFromFox(n)
		}
	case *readerFlag:
		download.GetFromReader(n)
	case *update:
		s.UpdateMangaLib()
	default:
		fmt.Println("Default source used: MangaReader.")
		download.GetFromReader(n)
	}

}
