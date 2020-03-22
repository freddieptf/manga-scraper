// +build deadsource

package mangastream

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

const (
	testManga = "one"
)

func TestGetMangaStreamMangaList(t *testing.T) {
	mangaList, err := getMangaStreamMangaList()
	if err != nil {
		t.Fatal(err)
	}
	if len(mangaList) <= 0 {
		t.Error("mangalist was empty")
	}
	for _, manga := range mangaList {
		if manga.MangaID == "" || manga.MangaID == "none" {
			t.Errorf("mangaID %s was invalid", manga.MangaID)
		}
	}
}

func TestSearchMangaStreamManga(t *testing.T) {
	results, err := seachMangaList(testManga)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) <= 0 {
		t.Error("search results were empty")
	}
	for _, manga := range results {
		if !strings.Contains(strings.ToLower(manga.MangaName), testManga) {
			t.Errorf("search results might be wrong: searched %s got %s", testManga, manga.MangaName)
		}
	}
	fmt.Printf("%v\n", results)
}

func TestGetMangaChapterList(t *testing.T) {
	results, err := seachMangaList(testManga)
	if err != nil {
		t.Fatal(err)
	}
	chapterList, err := getMangaChapterList(results[0].MangaID)
	if err != nil {
		t.Fatal(err)
	}
	if len(chapterList) <= 0 {
		t.Error("chapter list was empty")
	}
	for _, chapter := range chapterList {
		if _, err := url.ParseRequestURI(chapter.URL); err != nil {
			t.Errorf("chapter url %s was invalid\n", chapter.URL)
		}
		if chapter.MangaName == "" {
			t.Error("chapter mangaName was invalid")
		}
		if chapter.ID == "" {
			t.Error("chapter ID was invalid")
		}
	}
}

func TestGetMangaChapterPages(t *testing.T) {
	results, err := seachMangaList(testManga)
	if err != nil {
		t.Fatal(err)
	}
	chapterList, err := getMangaChapterList(results[0].MangaID)
	if err != nil {
		t.Fatal(err)
	}
	chPages, err := getChapterPages(chapterList[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	for _, chPage := range chPages {
		if _, err := url.ParseRequestURI(chPage); err != nil {
			t.Errorf("url invalid %s\n", err)
		}
	}
}

func TestGetImageURLFromChapterPage(t *testing.T) {
	results, err := seachMangaList(testManga)
	if err != nil {
		t.Fatal(err)
	}
	chapterList, err := getMangaChapterList(results[0].MangaID)
	if err != nil {
		t.Fatal(err)
	}
	chPages, err := getChapterPages(chapterList[0].URL)
	if err != nil {
		t.Fatal(err)
	}

	imgURL, err := getImgURLFromChapterPage(chPages[0])
	if err != nil {
		t.Fatal(err)
	}
	if _, err := url.ParseRequestURI(imgURL); err != nil {
		t.Errorf("imgURL invalid %s\n", imgURL)
	}
	fmt.Printf("%v\n", imgURL)
}
