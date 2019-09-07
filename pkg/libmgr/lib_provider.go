package libmgr

import (
	"errors"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"log"
	"strconv"
)

type LibraryProvider interface {
	GetManga() (map[SourceLibProvider][]scraper.Manga, error)
	GetLatestChapter(source SourceLibProvider, manga scraper.Manga) (scraper.Chapter, error)
	GetSourceManga(sourceLibProvider SourceLibProvider) ([]scraper.Manga, error)
}

type SourceLibProvider interface {
	Name() string
	Search(mangaName string) ([]scraper.Manga, error)
	ListMangaChapters(mangaID string) ([]scraper.Chapter, error)
	GetChapter(mangaID, chapterID string) (scraper.Chapter, error)
}

type LibraryManager struct{}

func (*LibraryManager) GetLibrarySourceUpdates(libProvider LibraryProvider, sourceLibProvider SourceLibProvider) map[scraper.Manga][]scraper.Chapter {
	updates := make(map[scraper.Manga][]scraper.Chapter)
	mangas, err := libProvider.GetSourceManga(sourceLibProvider)
	if err != nil {
		log.Fatal(err)
	}
	for _, manga := range mangas {
		chUpdates, err := getMangaUpdates(libProvider, sourceLibProvider, &manga)
		if err != nil {
			log.Printf("%v\n", err)
		} else {
			updates[manga] = *chUpdates
		}
	}
	return updates
}

func (*LibraryManager) GetLibraryUpdates(libProvider LibraryProvider) map[scraper.Manga][]scraper.Chapter {
	updates := make(map[scraper.Manga][]scraper.Chapter)
	lib, err := libProvider.GetManga()
	if err != nil {
		log.Fatal(err)
	}
	for sourceProvider, mangas := range lib {
		for _, manga := range mangas {
			chUpdates, err := getMangaUpdates(libProvider, sourceProvider, &manga)
			if err != nil {
				log.Printf("%v\n", err)
			} else {
				updates[manga] = *chUpdates
			}
		}
	}
	return updates
}

func (*LibraryManager) GetMangaUpdates(libProvider LibraryProvider, sourceProvider SourceLibProvider, manga *scraper.Manga) (*[]scraper.Chapter, error) {
	return getMangaUpdates(libProvider, sourceProvider, manga)
}

func getMangaUpdates(libProvider LibraryProvider, sourceLibProvider SourceLibProvider, manga *scraper.Manga) (*[]scraper.Chapter, error) {
	chapter, err := libProvider.GetLatestChapter(sourceLibProvider, *manga)
	if err != nil {
		log.Printf("couldn't get latest chapter %v\n", err)
	} else {
		chapters, err := getLatestChapters(sourceLibProvider, manga, chapter.ID)
		if err != nil {
			log.Printf("couldn't get latest chapters %v\n", err)
		} else {
			return &chapters, nil
		}
	}
	return nil, errors.New("No Updates")
}

func getLatestChapters(libProvider SourceLibProvider, manga *scraper.Manga, fromChapterID string) ([]scraper.Chapter, error) {
	if manga.MangaID == "" {
		results, err := libProvider.Search(manga.MangaName)
		if err != nil {
			return []scraper.Chapter{}, err
		}
		*manga = results[0]
	}
	chapters, err := libProvider.ListMangaChapters(manga.MangaID)
	if err != nil {
		return []scraper.Chapter{}, err
	}
	results := []scraper.Chapter{}
	for _, ch := range chapters {
		chID, err := strconv.ParseFloat(ch.ID, 64)
		if err != nil {
			log.Fatal(err)
		}
		fromChID, err := strconv.ParseFloat(fromChapterID, 64)
		if err != nil {
			log.Fatal(err)
		}
		if chID > fromChID {
			results = append(results, ch)
		}
	}
	return results, nil
}
