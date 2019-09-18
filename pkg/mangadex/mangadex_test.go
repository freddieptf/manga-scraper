package mangadex

import (
	"context"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetMangaDexMangaListingPageLinks(t *testing.T) {
	pageLinks, err := GetMangaDexListingPageLinks()
	if err != nil {
		t.Fatal(err)
	}
	// why 100? no reason really..mangadex is huge so this is probably a safe
	// guess regardless of no of manga entries per page
	if len(pageLinks) < 100 {
		t.Fail()
	}
}

func TestGetList(t *testing.T) {
	pageLinks, err := GetMangaDexListingPageLinks()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	resultChan := ScrapeMangaDexListing(ctx, 5, pageLinks[:2])
	mangas := []scraper.Manga{}
	for i := len(pageLinks[:2]); i > 0; i-- {
		result := <-resultChan
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		mangas = append(mangas, result.Mangas...)
	}
	if len(mangas) <= 0 {
		t.Fail()
	}
}

func TestCacheMangaList(t *testing.T) {
	mangas := &[]scraper.Manga{scraper.Manga{MangaName: "Test", MangaID: "Test"}}
	cacheFilePath := filepath.Join(os.Getenv("HOME"), ".mangadex-scraper-cache-test.txt")
	err := cacheMangaListToFS(cacheFilePath, mangas)
	if err != nil {
		t.Fatal(err)
	}
	cacheMangas, err := readMangaListFromFSCache(cacheFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(*cacheMangas) != len(*mangas) {
		t.Fail()
	}
	if err := os.Remove(cacheFilePath); err != nil {
		t.Fatal(err)
	}
}

func TestGetMangaChapterList(t *testing.T) {
	chapters, err := getMangaChapterList("7139")
	if err != nil {
		t.Fatal(err)
	}
	if len(chapters) <= 0 {
		t.Fail()
	}
}

func TestGetMangaChapter(t *testing.T) {
	chapter, err := getMangaChapter("7139", "1")
	if err != nil {
		t.Fatal(err)
	}
	if len(chapter.ChapterPages) <= 0 {
		t.Fail()
	}
	for _, page := range chapter.ChapterPages {
		if _, err := url.ParseRequestURI(page.Url); err != nil {
			t.Errorf("url invalid %s\n", err)
		}
	}
}
