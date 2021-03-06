package cli

import (
	"context"
	"fmt"
	"github.com/freddieptf/manga-scraper/pkg/libmgr"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	scraper "github.com/freddieptf/manga-scraper/pkg/scraper"
)

type MangaSource interface {
	Name() string
	Search(mangaName string) ([]scraper.Manga, error)
	ListMangaChapters(mangaID string) ([]scraper.Chapter, error)
	GetChapter(mangaID, chapterID string) (scraper.Chapter, error)
}

func Get(n int, mangaName string, chapters *[]int, archive *bool, source MangaSource) {
	results, err := source.Search(mangaName)
	if err != nil {
		log.Fatal(err)
	}
	result := GetMatchFromSearchResults(ReadWrite{os.Stdin, os.Stdout}, results)

	downloader := initDownloader(context.TODO(), n)
	downloader.initWorkers()

	for _, chID := range *chapters {
		ch, err := source.GetChapter(result.MangaID, strconv.Itoa(chID))
		if err != nil {
			log.Printf("getChapter %d, err: %v\n", chID, err)
		} else {
			ch.MangaName = result.MangaName
			downloader.queueDownload(&ch, *archive)
		}
	}

	downloader.waitTillQueueEmptyAndExit()
}

func UpdateSourceManga(source libmgr.SourceLibProvider, manga string, n int, archive bool) {
	update(source, manga, n, archive)
}

func UpdateSourceLibrary(source libmgr.SourceLibProvider, n int, archive bool) {
	update(source, "", n, archive)
}

func update(source libmgr.SourceLibProvider, mangaName string, n int, archive bool) {
	downloader := initDownloader(context.TODO(), n)
	downloader.initWorkers()

	fsProvider := &LocalFSLibProvider{libraryRootPath: filepath.Join(os.Getenv("HOME"), "Manga")}
	libManager := &libmgr.LibraryManager{}

	if mangaName == "" {
		updates := libManager.GetLibrarySourceUpdates(fsProvider, source)
		fmt.Printf("%v\n", updates)
		for manga, chapters := range updates {
			for _, chi := range chapters {
				ch, err := source.GetChapter(manga.MangaID, chi.ID)
				if err != nil {
					log.Printf("err: owwie %s\n", err)
				} else {
					ch.MangaName = manga.MangaName
					downloader.queueDownload(&ch, archive)
				}
			}
		}

	} else {
		fsResults, err := fsProvider.GetSourceManga(source)
		if err != nil {
			log.Fatal(err)
		}
		results := []scraper.Manga{}
		for _, s := range fsResults {
			if strings.Contains(strings.ToLower(s.MangaName), strings.ToLower(mangaName)) {
				results = append(results, s)
			}
		}
		manga := GetMatchFromSearchResults(ReadWrite{os.Stdin, os.Stdout}, results)
		updates, _ := libManager.GetMangaUpdates(fsProvider, source, &manga)
		fmt.Printf("%v\n", updates)
		for _, chi := range *updates {
			ch, err := source.GetChapter(manga.MangaID, chi.ID)
			if err != nil {
				log.Printf("err: owwie %s\n", err)
			} else {
				ch.MangaName = manga.MangaName
				downloader.queueDownload(&ch, archive)
			}
		}
	}

	downloader.waitTillQueueEmptyAndExit()
}
