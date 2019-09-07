package main

import (
	"log"
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/mangastream"
	"os"
)

var (
	archive            bool
	maxActiveDownloads int 
	manga string
)

func main() {
	cmdFlagSet := flag.NewFlagSet("get", flag.ExitOnError)
	cmdFlagSet.StringVar(&manga, "manga", "", "manga name")
	cmdFlagSet.IntVar(&maxActiveDownloads, "n", 1, "max number of parallel downloads")
	cmdFlagSet.BoolVar(&archive, "cbz", true, "save as cbz file format")

	osArgs := os.Args
	if len(osArgs) <= 1 {
		cmdFlagSet.Usage()
		os.Exit(1)
	}

	switch osArgs[1] {
	case "get":
		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		if manga == "" {
			fmt.Println("no manga name was passed")
			cmdFlagSet.Usage()
			os.Exit(1)
		}
		args := cmdFlagSet.Args()
		source := &mangastream.MangaStream{}
		cli.Get(maxActiveDownloads, manga, cli.GetChapterRangeFromArgs(&args), &archive, source)
	}
}
