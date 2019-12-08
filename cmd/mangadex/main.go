package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	cli "github.com/freddieptf/manga-scraper/pkg/cli"
	"github.com/freddieptf/manga-scraper/pkg/mangadex"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	archive            bool
	maxActiveDownloads int
	manga              string
	cacheFilePath      string
)

func main() {
	getCmdFlagSet := flag.NewFlagSet("get", flag.ExitOnError)
	getCmdFlagSet.StringVar(&manga, "manga", "", "manga name")
	getCmdFlagSet.StringVar(&cacheFilePath, "cache", filepath.Join(os.Getenv("HOME"), ".mangadex-cache.txt"), "path of the cache file")
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
		updateCmdFlagSet.Usage()
		os.Exit(1)
	}

	switch osArgs[1] {
	case "get":
		{
			if osArgs[2] == "listing" {
				if err := getCmdFlagSet.Parse(os.Args[3:]); err != nil {
					log.Fatal(err)
				}
				getMangaDexListing(maxActiveDownloads)
			} else {
				if err := getCmdFlagSet.Parse(os.Args[2:]); err != nil {
					log.Fatal(err)
				}
				source := &mangadex.MangaDex{CachePath: cacheFilePath}
				if manga == "" {
					fmt.Println("no manga name was passed")
					getCmdFlagSet.Usage()
					os.Exit(1)
				}
				args := getCmdFlagSet.Args()
				cli.Get(maxActiveDownloads, manga, cli.GetChapterRangeFromArgs(&args), &archive, source)
			}
		}
	case "update":
		{
			if err := updateCmdFlagSet.Parse(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			source := &mangadex.MangaDex{CachePath: cacheFilePath}
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

func getMangaDexListing(workers int) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGQUIT, syscall.SIGINT)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(cancelFunc func()) {
		<-signalChan
		fmt.Fprintf(os.Stderr, "\nreceived user interrupt..trying to stop gracefully\n")
		cancelFunc()
	}(cancelFunc)

	listingPageLinks, err := mangadex.GetMangaDexListingPageLinks()
	if err != nil {
		log.Fatal(err)
	}
	defer cancelFunc()
	resultChan := mangadex.ScrapeMangaDexListing(ctx, workers, listingPageLinks)
	mangas := []scraper.Manga{}
	defer writeOutMangas(&mangas)
	n := len(listingPageLinks)
	for result := range resultChan {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "couldn't get what we wanted: %s", result.Err)
		} else {
			mangas = append(mangas, result.Mangas...)
		}
		n--
		if n <= 0 {
			return
		}
	}
}

func writeOutMangas(mangas *[]scraper.Manga) {
	mangasjson, err := json.Marshal(mangas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "oops %v\n", err)
	} else {
		fmt.Printf("%s", mangasjson)
	}
}
