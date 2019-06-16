package mangareader

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
)

const (
	mangaReaderURL = "http://www.mangareader.net"
)

type ReaderManga struct{}

// pass in the document we get when we load the first chapter page and try and get all the
// other chapter pages urls from the select component
func getChPageUrls(doc *goquery.Document) (chapterPageUrls []string) {
	doc.Find("div#selectpage > select#pageMenu > option").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("value")
		url = mangaReaderURL + url
		chapterPageUrls = append(chapterPageUrls, url)
	})
	return
}

// pass in a chapter page url and try and get the url of the image in the page
func getChPageImgUrl(url string) (imgUrl string) {
	doc, err := scraper.MakeDocRequest(url)
	if err != nil {
		log.Println(err)
		return
	}
	imgUrl, _ = doc.Find("div#imgholder img").Attr("src")
	if imgUrl == "" {
		log.Printf("weird..couldn't get chapter image url from %s\n", url)
	}
	return
}

func getChTitle(mangaID, chapterID string) (title string) {
	d, e := scraper.MakeDocRequest(mangaReaderURL + mangaID)
	if e != nil {
		log.Println(e)
		return
	}
	d.Find("div#chapterlist > table#listing td").Has("div.chico_manga").EachWithBreak(func(i int, s *goquery.Selection) bool {
		link, _ := s.Find("a").Attr("href")
		ps := strings.Split(link, "/")
		if ps[2] == chapterID {
			title = strings.Split(s.Text(), " : ")[1]
			return false
		}
		return true
	})
	return
}

func (readerManga *ReaderManga) GetChapter(mangaID, chapterID string) (scraper.Chapter, error) {
	var (
		chapterPageUrls []string
		chapterPages    []scraper.ChapterPage
		doc             *goquery.Document
		err             error
	)

	chapterUrl := mangaReaderURL + mangaID + "/" + chapterID

	doc, err = scraper.MakeDocRequest(chapterUrl)
	if err != nil {
		return scraper.Chapter{}, err
	}

	chapterPageUrls = getChPageUrls(doc)

	// go off and get the chapter title from the chapter listings
	titleChan := make(chan string)
	go func() {
		titleChan <- getChTitle(mangaID, chapterID)
	}()

	//scrape the manga pages for the image urls
	log.Printf("%s: Getting the chapter image urls\n", chapterUrl)

	pageChan := make(chan scraper.ChapterPage)
	for i, url := range chapterPageUrls {
		go func(i int, url string) {
			pageChan <- scraper.ChapterPage{Page: i, Url: getChPageImgUrl(url)}
		}(i, url)
	}

	for i := 0; i < len(chapterPageUrls); i++ {
		chapterPages = append(chapterPages, <-pageChan)
	}

	chapterTitle := <-titleChan
	chapter := scraper.Chapter{
		MangaName:    "",
		ChapterTitle: fmt.Sprintf("%s: %s", chapterID, chapterTitle),
		ChapterPages: chapterPages,
		SourceName:   "MangaReader",
	}

	return chapter, nil
}

func (readerManga *ReaderManga) Search(mangaName string) ([]scraper.Manga, error) {
	doc, err := scraper.MakeDocRequest(mangaReaderURL + "/alphabetical")
	results := []scraper.Manga{}
	if err != nil {
		return results, err
	}
	//find possible matches in the site's manga list for the mangaName provided;
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(mangaName)) {
			mid, _ := s.Attr("href")
			results = append(results, scraper.Manga{MangaName: s.Text(), MangaID: mid})
		}
	})

	if len(results) <= 0 {
		return results, errors.New("found Zero results")
	}

	return results, nil
}
