package main

import (
	s "GoManga/source"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	var (
		start, stop int
		err         error
	)

	chapters := os.Args[1:]

	switch len(chapters) {
	case 2:
		start, err = strconv.Atoi(chapters[1])
		if err != nil {
			log.Fatal(err)
		}
		s.Get(start, -1, chapters[0])
	case 3:
		start, err = strconv.Atoi(chapters[1])
		if err != nil {
			log.Fatal(err)
		}
		stop, err = strconv.Atoi(chapters[2])
		if err != nil {
			log.Fatal(err)
		}
		s.Get(start, stop, chapters[0])
	default:
		fmt.Println("Usage Format: 'Manga Name' startChapter endChapter(optional)")
		fmt.Println("Usage Format: 'Bleach' 1 50 (to fetch chapter 1 to 50)")
	}

}
