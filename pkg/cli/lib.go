package cli

import (
	"fmt"
	"github.com/freddieptf/manga-scraper/pkg/libmgr"
	"github.com/freddieptf/manga-scraper/pkg/mangareader"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type LocalFSLibProvider struct {
	libraryRootPath string
}

func getMangaSources() map[string]libmgr.SourceLibProvider {
	sources := make(map[string]libmgr.SourceLibProvider)
	mangaReader := &mangareader.ReaderManga{}
	sources[mangaReader.Name()] = mangaReader
	return sources
}

func getLocalSourceLibraryManga(libraryRootPath string, srcLibProvider libmgr.SourceLibProvider) ([]scraper.Manga, error) {
	rootMangaLibDir, err := os.Open(libraryRootPath)
	if err != nil {
		return nil, err
	}
	sourceDirNames, err := rootMangaLibDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	for _, sourceName := range sourceDirNames {
		if sourceName == srcLibProvider.Name() {
			sourceManga, err := getSourceManga(libraryRootPath, sourceName)
			if err != nil {
				return nil, err
			}
			return *sourceManga, nil
		}
	}
	return []scraper.Manga{}, nil
}

func getLocalLibraryMap(libraryRootPath string) (map[libmgr.SourceLibProvider][]scraper.Manga, error) {
	rootMangaLibDir, err := os.Open(libraryRootPath)
	if err != nil {
		return nil, err
	}
	sourceDirNames, err := rootMangaLibDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sources := getMangaSources()
	libMap := make(map[libmgr.SourceLibProvider][]scraper.Manga)
	for _, sourceName := range sourceDirNames {
		if _, exists := sources[sourceName]; exists {
			sourceManga, err := getSourceManga(libraryRootPath, sourceName)
			if err != nil {
				return nil, err
			}
			libMap[sources[sourceName]] = *sourceManga
		}
	}
	return libMap, nil
}

func getSourceManga(libraryRootPath, sourceName string) (*[]scraper.Manga, error) {
	sourceDir, err := os.Open(filepath.Join(libraryRootPath, sourceName))
	if err != nil {
		return nil, err
	}
	mangaNames, err := sourceDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sourceManga := []scraper.Manga{}
	for _, mangaName := range mangaNames {
		sourceManga = append(sourceManga, scraper.Manga{MangaName: mangaName})
	}
	return &sourceManga, nil
}

func getLocalLatestMangaChapter(libraryRootPath string, mangaSource MangaSource, manga scraper.Manga) (scraper.Chapter, error) {
	chapterTitles := []string{}
	err := filepath.Walk(filepath.Join(libraryRootPath, mangaSource.Name(), manga.MangaName),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			chapterTitles = append(chapterTitles, info.Name())
			if info.IsDir() && info.Name() != manga.MangaName {
				return filepath.SkipDir
			} else {
				return nil
			}
		})
	if err != nil {
		return scraper.Chapter{}, err
	}

	chapterIDs := []float64{}
	for _, title := range chapterTitles {
		chs := strings.SplitAfterN(title, "-", 2)
		f, err := strconv.ParseFloat(strings.TrimSuffix(chs[0], " -"), 64)
		if err != nil {
			log.Println(err)
		} else {
			chapterIDs = append(chapterIDs, f)
		}
	}
	sort.Float64s(chapterIDs)
	return scraper.Chapter{
		ID:         fmt.Sprintf("%f", chapterIDs[len(chapterIDs)-1]),
		MangaName:  manga.MangaName,
		SourceName: mangaSource.Name(),
	}, nil
}

func (provider *LocalFSLibProvider) GetManga() (map[libmgr.SourceLibProvider][]scraper.Manga, error) {
	return getLocalLibraryMap(provider.libraryRootPath)
}

func (provider *LocalFSLibProvider) GetSourceManga(sourceLibProvider libmgr.SourceLibProvider) ([]scraper.Manga, error) {
	return getLocalSourceLibraryManga(provider.libraryRootPath, sourceLibProvider)
}

func (provider *LocalFSLibProvider) GetLatestChapter(source libmgr.SourceLibProvider, manga scraper.Manga) (scraper.Chapter, error) {
	return getLocalLatestMangaChapter(provider.libraryRootPath, source, manga)
}
