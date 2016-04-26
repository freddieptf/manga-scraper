package main

import (
	fox "GoManga/mangafox"
	reader "GoManga/mangareader"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	chapters := os.Args[1:]

	switch len(chapters) {
	case 2:
		start, _ := getStartStop(chapters[1], "")
		fmt.Println("Default source used: MangaReader.")
		reader.Get(start, -1, chapters[0])
	case 3:
		if strings.Contains(chapters[0], "-mr") || strings.Contains(chapters[0], "-mf") {
			start, _ := getStartStop(chapters[2], "")
			if strings.Compare(strings.ToLower(chapters[0]), "-mr") == 0 {
				reader.Get(start, -1, chapters[1])
			} else if strings.Compare(strings.ToLower(chapters[0]), "-mf") == 0 {
				fox.Get(start, -1, chapters[1])
			}
		} else {
			if _, err := strconv.Atoi(chapters[1]); err == nil {
				start, stop := getStartStop(chapters[1], chapters[2])
				fmt.Println("Default source used: MangaReader.")
				reader.Get(start, stop, chapters[0])
			} else {
				printUsageFormat()
			}
		}
	case 4:
		if strings.Compare(strings.ToLower(chapters[0]), "-mr") == 0 {
			start, stop := getStartStop(chapters[2], chapters[3])
			reader.Get(start, stop, chapters[1])
		} else if strings.Compare(strings.ToLower(chapters[0]), "-mf") == 0 {
			start, stop := getStartStop(chapters[2], chapters[3])
			fox.Get(start, stop, chapters[1])
		} else {
			printUsageFormat()
			break
		}
	default:
		printUsageFormat()
	}

}

func getStartStop(start, stop string) (int, int) {
	var x, y int
	var err error
	x, err = strconv.Atoi(start)
	if err != nil {
		log.Fatal(err)
	}
	if stop != "" {
		y, err = strconv.Atoi(stop)
		if err != nil {
			log.Fatal(err)
		}
	}
	return x, y
}

func printUsageFormat() {
	fmt.Printf("\nUsage Format: \n\t-source 'Manga Name' start stop(optional)\n\n")
	fmt.Printf("sources: \n\t-mr for mangareader \n\t-mf for mangafox\n\n")
	fmt.Println("To fetch Bleach, chapter 1 to 50, from mangareader: ")
	fmt.Printf("\t sudo ./GoManga -mr 'Bleach' 1 50\n\n")
}
