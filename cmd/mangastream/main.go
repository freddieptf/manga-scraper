// +build deadsource

package main

import (
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/mangastream"
	"log"
	"os"
)

var (
	archive            bool
	maxActiveDownloads int
	manga              string
)

func main() {
	getCmdFlagSet := flag.NewFlagSet("get", flag.ExitOnError)
	getCmdFlagSet.StringVar(&manga, "manga", "", "manga name")
	getCmdFlagSet.IntVar(&maxActiveDownloads, "n", 1, "max number of parallel downloads")
	getCmdFlagSet.BoolVar(&archive, "cbz", false, "save as cbz file format")

	updateCmdFlagSet := flag.NewFlagSet("update", flag.ExitOnError)
	updateCmdFlagSet.StringVar(&manga, "manga", "", "update this manga")
	updateAll := updateCmdFlagSet.Bool("all", false, "update all")
	updateCmdFlagSet.IntVar(&maxActiveDownloads, "n", 1, "max number of parallel downloads")
	updateCmdFlagSet.BoolVar(&archive, "cbz", false, "save as cbz file format")

	osArgs := os.Args
	if len(osArgs) <= 1 {
		getCmdFlagSet.Usage()
		os.Exit(1)
	}

	switch osArgs[1] {
	case "get":
		{
			if err := getCmdFlagSet.Parse(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			if manga == "" {
				fmt.Println("no manga name was passed")
				getCmdFlagSet.Usage()
				os.Exit(1)
			}
			args := getCmdFlagSet.Args()
			source := &mangastream.MangaStream{}
			cli.Get(maxActiveDownloads, manga, cli.GetChapterRangeFromArgs(&args), &archive, source)
		}
	case "update":
		{
			if err := updateCmdFlagSet.Parse(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			source := &mangastream.MangaStream{}
			if *updateAll {
				cli.UpdateSourceLibrary(source, maxActiveDownloads, archive)
			} else {
				if manga != "" {
					cli.UpdateSourceManga(source, manga, maxActiveDownloads, archive)
				} else {
					updateCmdFlagSet.Usage()
				}
			}
		}
	}
}
