package main

import (
	s "GoManga/msources"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	foxFlag            = kingpin.Flag("mf", "use mangafox as source").Bool()
	readerFlag         = kingpin.Flag("mr", "use mangaReader as source").Bool()
	v                  = kingpin.Flag("vlm", "use when you want to download a volume(s)").Bool()
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
		Chapters:  getRange(args),
		MangaName: manga,
	}

	switch {
	case *foxFlag:
		download.GetFromFox(n)
	case *readerFlag:
		download.GetFromReader(n)
	default:
		fmt.Println("Default source used: MangaReader.")
		download.GetFromReader(n)
	}

}

func getRange(vals *[]string) *[]int {
	var x, y int
	var err error
	var chapters []int
	for _, val := range *vals {
		if strings.Contains(val, "-") {
			chs := strings.Split(val, "-")
			x, err = strconv.Atoi(chs[0])
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			y, err = strconv.Atoi(chs[1])
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			for i := x; i <= y; i++ {
				chapters = append(chapters, i)
			}
		} else {
			x, err = strconv.Atoi(val)
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			chapters = append(chapters, x)
		}
	}

	fmt.Printf("%v\n", chapters)

	return &chapters
}
