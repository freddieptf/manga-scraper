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
	maxActiveDownloads = kingpin.Flag("n", "set max number of concurrent downloads").Int()
	manga              = kingpin.Arg("manga", "The name of the manga").String()
	args               = kingpin.Arg("arguments",
		"chapters (volumes if --vlm is set) to download. Example format: 1 3 5-7 67 10-14").Strings()
)

func main() {
	kingpin.Parse()
	n := 35 //default num of maxActiveDownloads
	if *maxActiveDownloads != 0 {
		n = *maxActiveDownloads
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
	default:
		fmt.Println("Default source used: MangaReader.")
		download.GetFromReader(n)
	}

}
