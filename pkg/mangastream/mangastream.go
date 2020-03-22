// +build deadsource

package mangastream

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"strconv"
	"strings"
)

const (
	mangaStreamURL = "https://readms.net"
	sourceName     = "MangaStream"
)

type MangaStream struct{}

func getMangaStreamMangaList() ([]scraper.Manga, error) {
	doc, err := scraper.MakeDocRequest(fmt.Sprintf("%s/%s", mangaStreamURL, "manga"))
	if err != nil {
		return []scraper.Manga{}, err
	}
	results := []scraper.Manga{}
	doc.Find("table .table, .table-striped tr > td > strong > a").Each(func(i int, s *goquery.Selection) {
		mangaID := s.AttrOr("href", "none")
		results = append(results,
			scraper.Manga{
				MangaName: s.Text(),
				MangaID:   strings.TrimPrefix(mangaID, "/manga/"),
			})
	})

	return results, nil
}

func seachMangaList(mangaName string) ([]scraper.Manga, error) {
	mangaList, err := getMangaStreamMangaList()
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
	doc, err := scraper.MakeDocRequest(fmt.Sprintf("%s/manga/%s", mangaStreamURL, mangaID))
	if err != nil {
		return []scraper.Chapter{}, err
	}
	chapters := []scraper.Chapter{}
	mangaName := doc.Find("div.col-sm-8 > h1").First().Text()
	doc.Find("table .table, .table-striped tr > td > a").Each(func(i int, s *goquery.Selection) {
		chapterRelativeUrl := s.AttrOr("href", "none")
		var chURL, chID string
		if chapterRelativeUrl != "none" {
			paths := strings.SplitAfterN(strings.Trim(chapterRelativeUrl, "/"), "/", 4)
			chID = strings.Trim(paths[2], "/")
			chURL = fmt.Sprintf("%s%s", mangaStreamURL, chapterRelativeUrl)
		}
		chapters = append(chapters, scraper.Chapter{
			MangaName:    strings.Trim(mangaName, " \n"),
			ChapterTitle: strings.Trim(s.Text(), " "),
			SourceName:   sourceName,
			URL:          chURL,
			ID:           chID,
		})
	})
	return chapters, nil
}

func getChapterPages(chapterUrl string) ([]string, error) {
	firstPageDoc, err := scraper.MakeDocRequest(chapterUrl)
	if err != nil {
		return nil, err
	}

	basePagePaths := strings.Split(strings.TrimPrefix(chapterUrl, mangaStreamURL), "/")
	basePageUrl := strings.Join(basePagePaths[0:len(basePagePaths)-1], "/")

	lastPageURL := firstPageDoc.Find("div.btn-reader-page ul.dropdown-menu > li > a").Last().AttrOr("href", "none")
	paths := strings.Split(lastPageURL, "/")
	lastPageIdx, err := strconv.Atoi(paths[len(paths)-1])
	if err != nil {
		return nil, err
	}

	chPages := []string{}
	for i := 1; i <= lastPageIdx; i++ {
		chPages = append(chPages, fmt.Sprintf("%s%s/%d", mangaStreamURL, basePageUrl, i))
	}

	return chPages, nil
}

func getImgURLFromChapterPage(chapterPageURL string) (string, error) {
	doc, err := scraper.MakeDocRequest(chapterPageURL)
	if err != nil {
		return "", err
	}
	imgURL := doc.Find("div.page > a > img").First().AttrOr("src", "none")
	return fmt.Sprintf("https:%s", imgURL), nil
}

func (*MangaStream) Name() string {
	return sourceName
}

func (*MangaStream) ListMangaDirectory() ([]scraper.Manga, error) {
	return getMangaStreamMangaList()
}

func (*MangaStream) Search(mangaName string) ([]scraper.Manga, error) {
	return seachMangaList(mangaName)
}

func (*MangaStream) ListMangaChapters(mangaID string) ([]scraper.Chapter, error) {
	return getMangaChapterList(mangaID)
}

func (*MangaStream) GetChapter(mangaID, chapterID string) (scraper.Chapter, error) {
	chapterList, err := getMangaChapterList(mangaID)
	if err != nil {
		return scraper.Chapter{}, err
	}

	for _, chapter := range chapterList {
		// lol ok
		chInputAsFloat, _ := strconv.ParseFloat(chapterID, 32)
		chAsFloat, _ := strconv.ParseFloat(chapter.ID, 32)
		if chAsFloat == chInputAsFloat {
			chapterPageURLS, err := getChapterPages(chapter.URL)
			if err != nil {
				return scraper.Chapter{}, err
			}
			chapterPages := []scraper.ChapterPage{}
			for i, pageUrls := range chapterPageURLS {
				imgURL, err := getImgURLFromChapterPage(pageUrls)
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
