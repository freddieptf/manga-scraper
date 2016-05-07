package main

import (
	s "GoManga/msources"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	var download s.MangaDownload
	n := 35 //default num of maxActiveDownloads

	foxFlag := flag.Bool("mf", false, "use mangafox as source")
	readerFlag := flag.Bool("mr", false, "use mangaReader as source")
	maxActiveDownloads := flag.Int("n", -1, "Define the maximum number of concurrent downloads")
	flag.Parse()

	if len(args) > 1 {
		switch {
		case *foxFlag, *readerFlag:
			if *maxActiveDownloads != -1 {
				download = s.MangaDownload{
					Chapters:  getChapterRange(args[3:]),
					MangaName: &args[2],
				}
				n = *maxActiveDownloads
			} else {
				download = s.MangaDownload{
					Chapters:  getChapterRange(args[2:]),
					MangaName: &args[1],
				}
			}

			if *readerFlag {
				download.GetFromReader(n)
			} else if *foxFlag {
				download.GetFromFox(n)
			}

		default:
			if *maxActiveDownloads != -1 {
				download = s.MangaDownload{
					Chapters:  getChapterRange(args[2:]),
					MangaName: &args[1],
				}
				n = *maxActiveDownloads
			} else {
				download = s.MangaDownload{
					Chapters:  getChapterRange(args[1:]),
					MangaName: &args[0],
				}
			}
			fmt.Println("Default source used: MangaReader.")
			download.GetFromReader(n)
		}

	} else {
		printUsageFormat()
	}

}

func getChapterRange(vals []string) *[]int {
	var x, y int
	var err error
	var chapters []int
	for _, val := range vals {
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

func printUsageFormat() {
	fmt.Printf("\nUsage Format: \n\t-source 'Manga Name' chapters\n\n")
	fmt.Printf("sources: \n\t-mr for mangareader \n\t-mf for mangafox\n\n")
	fmt.Println("To fetch Bleach, chapters 1 to 20 and chapter 54, from mangafox: ")
	if runtime.GOOS == "linux" {
		fmt.Printf("\t ./GoManga -mf 'Bleach' 1-20 54\n\n")
	} else if runtime.GOOS == "windows" {
		fmt.Printf("\t GoManga.exe -mf 'Bleach' 1-20 54\n\n")
	}
}
