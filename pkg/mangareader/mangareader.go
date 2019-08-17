package mangareader

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
)

const (
	mangaReaderURL = "http://www.mangareader.net"
	sourceName     = "MangaReader"
)

type ReaderManga struct{}

func getChPageUrls(chapterPageURL string) ([]string, error) {
	doc, err := scraper.MakeDocRequest(chapterPageURL)
	if err != nil {
		return []string{}, err
	}
	pageURLs := []string{}
	doc.Find("div#selectpage > select#pageMenu > option").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("value")
		url = mangaReaderURL + url
		pageURLs = append(pageURLs, url)
	})
	return pageURLs, nil
}

// pass in a chapter page url and try and get the url of the image in the page
func getChPageImgUrl(url string) (string, error) {
	doc, err := scraper.MakeDocRequest(url)
	if err != nil {
		return "", err
	}
	return doc.Find("div#imgholder img").AttrOr("src", "none"), nil
}

func getMangaReaderMangaList() ([]scraper.Manga, error) {
	results := []scraper.Manga{}
	doc, err := scraper.MakeDocRequest(mangaReaderURL + "/alphabetical")
	if err != nil {
		return results, err
	}
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		mid := strings.TrimPrefix(s.AttrOr("href", "none"), "/")
		results = append(results, scraper.Manga{MangaName: s.Text(), MangaID: mid})
	})
	return results, nil
}

func searchMangaList(mangaName string) ([]scraper.Manga, error) {
	mangaList, err := getMangaReaderMangaList()
	if err != nil {
		return []scraper.Manga{}, err
	}
	results := []scraper.Manga{}
	for _, manga := range mangaList {
		if strings.Contains(strings.ToLower(manga.MangaName), strings.ToLower(mangaName)) {
			results = append(results, manga)
		}
	}
	return results, nil
}

func getMangaChapterList(mangaID string) ([]scraper.Chapter, error) {
	results := []scraper.Chapter{}
	doc, err := scraper.MakeDocRequest(fmt.Sprintf("%s/%s", mangaReaderURL, mangaID))
	if err != nil {
		return results, err
	}
	mangaName := doc.Find("h2.aname").Text()
	doc.Find("div#chapterlist > table#listing td").Has("div.chico_manga").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		results = append(results, scraper.Chapter{
			ID:           strings.TrimPrefix(a.AttrOr("href", "none"), fmt.Sprintf("/%s/", mangaID)),
			ChapterTitle: strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSpace(s.Text()), strings.TrimSpace(a.Text())), ":", "")),
			SourceName:   sourceName,
			URL:          fmt.Sprintf("%s%s", mangaReaderURL, a.AttrOr("href", "none")),
			MangaName:    strings.TrimSpace(mangaName),
		})
	})
	return results, nil
}

func (readerManga *ReaderManga) GetChapter(mangaID, chapterID string) (scraper.Chapter, error) {
	chapterList, err := getMangaChapterList(mangaID)
	if err != nil {
		return scraper.Chapter{}, err
	}

	for _, chapter := range chapterList {
		chInputAsFloat, _ := strconv.ParseFloat(chapterID, 32)
		chAsFloat, _ := strconv.ParseFloat(chapter.ID, 32)
		if chAsFloat == chInputAsFloat {
			chapterPageURLS, err := getChPageUrls(chapter.URL)
			if err != nil {
				return scraper.Chapter{}, err
			}
			chapterPages := []scraper.ChapterPage{}
			for i, pageURL := range chapterPageURLS {
				imgURL, err := getChPageImgUrl(pageURL)
				if err != nil {
					return scraper.Chapter{}, err
				}
				chapterPages = append(chapterPages, scraper.ChapterPage{Page: i, Url: imgURL})
			}
			chapter.ChapterPages = chapterPages
			return chapter, nil
		}
	}

	return scraper.Chapter{}, errors.New("chapter not found")
}

func (readerManga *ReaderManga) Search(mangaName string) ([]scraper.Manga, error) {
	return searchMangaList(mangaName)
}

func (*ReaderManga) Name() string {
	return sourceName
}

func (*ReaderManga) ListMangaDirectory() ([]scraper.Manga, error) {
	return getMangaReaderMangaList()
}

func (*ReaderManga) ListMangaChapters(mangaID string) ([]scraper.Chapter, error) {
	return getMangaChapterList(mangaID)
}
