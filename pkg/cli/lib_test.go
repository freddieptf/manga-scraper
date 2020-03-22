package cli

import (
	"fmt"
	"github.com/freddieptf/manga-scraper/pkg/mangareader"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

var (
	TEST_DIR_PATH = filepath.Join(os.Getenv("HOME"), "MangaScraperTest")
)

func setUpEmptyTestLib(t *testing.T) (string, map[MangaSource][]scraper.Manga) {
	err := os.Mkdir(TEST_DIR_PATH, 0777)
	if err != nil {
		t.Fatal(err)
	}
	sources := make(map[string]MangaSource)
	mangaReader := &mangareader.ReaderManga{}
	sources[mangaReader.Name()] = mangaReader
	for key, source := range sources {
		mangaPath := filepath.Join(TEST_DIR_PATH, source.Name(), key)
		err = os.MkdirAll(mangaPath, 0777)
		if err != nil {
			t.Fatal(err)
		}
	}
	// lets add some none source dirs for fun
	for _, val := range []string{"NOPE", "NOPEE"} {
		sourcePath := filepath.Join(TEST_DIR_PATH, val)
		err = os.MkdirAll(sourcePath, 0777)
		if err != nil {
			t.Fatal(err)
		}
	}
	return TEST_DIR_PATH, map[MangaSource][]scraper.Manga{
		mangaReader: []scraper.Manga{scraper.Manga{MangaName: mangaReader.Name()}},
	}
}

func tearDownTestDir(t *testing.T) {
	err := os.RemoveAll(TEST_DIR_PATH)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLocalLib(t *testing.T) {
	dirPath, wantedResults := setUpEmptyTestLib(t)
	lib, err := getLocalLibraryMap(dirPath)
	if err != nil {
		t.Fatal(err)
	}
	for i, result := range lib {
		if !reflect.DeepEqual(result, wantedResults[i]) {
			t.Fail()
		}
	}
	tearDownTestDir(t)
}

func TestGetLocalLatestMangaChapter(t *testing.T) {
	libPath, libMap := setUpEmptyTestLib(t)
	pool := [][]float64{
		[]float64{3.6, 1, 1.5, 3.5, 20},
		[]float64{1, 1.5, 22, 3.5, 3.55},
		[]float64{3, 10, 1.5, 3.5, 2},
	}
	cmap := make(map[scraper.Manga][]float64)
	rand.Seed(time.Now().UnixNano())
	for source, mangas := range libMap {
		for _, manga := range mangas {
			mangaPath := filepath.Join(libPath, source.Name(), manga.MangaName)
			chapters := pool[rand.Intn(3)]
			cmap[manga] = chapters
			// shit test..
			for idx, val := range chapters {
				if idx%3 == 0 { // create file and emulate a chapter cbz
					_, err := os.Create(filepath.Join(mangaPath, fmt.Sprintf("%f - Title.cbz", val)))
					if err != nil {
						t.Fatal(err)
					}
				} else { // create a folder and emulate a chapter folder..lol
					err := os.MkdirAll(filepath.Join(mangaPath, fmt.Sprintf("%f - Title.cbz", val)), 0777)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
	for k, v := range libMap {
		for _, manga := range v {
			chapter, err := getLocalLatestMangaChapter(libPath, k, manga)
			if err != nil {
				t.Fatal(err)
			}
			chapters := cmap[manga]
			sort.Float64s(chapters)
			if chapter.ID != fmt.Sprintf("%f", chapters[len(chapters)-1]) {
				t.Fail()
			}
		}
	}
	tearDownTestDir(t)
}
