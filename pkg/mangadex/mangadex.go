package mangadex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

const (
	MANGADEX_URL     = "https://mangadex.org"
	MANGADEX_API_URL = MANGADEX_URL + "/api"
	SOURCE_NAME      = "MangaDex"
)

type MangaDex struct {
	CachePath string
}

type DexScrapeResult struct {
	Mangas []scraper.Manga
	Err    error
}

type MangaDexScraperManager struct {
	workerCount      int
	resultchan       *chan DexScrapeResult
	workerDoneChan   chan struct{}
	aliveWorkerCount int
}

type mangaDexListScraper struct {
	workchan       *chan string
	resultchan     *chan DexScrapeResult
	ctx            context.Context
	workerDoneChan *chan struct{} // gonna need this to notify MangaDexScraperManager when this gets done
}

// need to know when this fails idiot
func scrapeMangaFromDoc(doc *goquery.Document) *[]scraper.Manga {
	manga := []scraper.Manga{}
	doc.Find("div.manga-entry").Each(func(idx int, s *goquery.Selection) {
		manga = append(manga, scraper.Manga{
			MangaName: s.Find("a.manga_title").AttrOr("title", "NONE"),
			MangaID:   s.AttrOr("data-id", "NONE")})
	})
	return &manga
}

func (*mangaDexListScraper) getMangaListingFromPage(url string) (*[]scraper.Manga, error) {
	doc, err := scraper.MakeDocRequest(url)
	if err != nil {
		return nil, err
	}
	mangas := scrapeMangaFromDoc(doc)
	return mangas, nil
}

func (m *mangaDexListScraper) standBy() {
	for {
		select {
		case url := <-*m.workchan:
			mangas, err := m.getMangaListingFromPage(url)
			*m.resultchan <- DexScrapeResult{Mangas: *mangas, Err: err}

		case <-m.ctx.Done():
			*m.workerDoneChan <- struct{}{}
			return
		}
	}
}

func (manager *MangaDexScraperManager) dispatchWorkers(ctx context.Context, workchan *chan string, resultChan *chan DexScrapeResult) {
	manager.workerDoneChan = make(chan struct{})
	for index := 0; index < manager.workerCount; index++ {
		worker := &mangaDexListScraper{workchan: workchan, resultchan: resultChan, ctx: ctx, workerDoneChan: &manager.workerDoneChan}
		go worker.standBy()
		manager.aliveWorkerCount++
	}
}

func (manager *MangaDexScraperManager) waitForWorkersDeath() {
	for {
		select {
		case <-manager.workerDoneChan:
			manager.aliveWorkerCount--
			if manager.aliveWorkerCount == 0 {
				close(*manager.resultchan)
				return
			}
		}
	}
}

func GetMangaDexListingPageLinks() ([]string, error) {
	doc, err := scraper.MakeDocRequest(fmt.Sprintf("%s/titles/2", MANGADEX_URL))
	if err != nil {
		return []string{}, err
	}
	last := doc.Find("nav > ul.pagination > li.page-item > a.page-link").Last().AttrOr("href", "BUTTS")
	if last == "BUTTS" {
		log.Fatal("couldn't find last manga list webpage link")
	}
	lastPage, err := strconv.Atoi(strings.TrimSuffix(strings.SplitAfterN(last, "/", 4)[3], "/"))
	if err != nil {
		log.Fatal(err)
	}
	urls := []string{}
	for index := 1; index <= lastPage; index++ {
		urls = append(urls, fmt.Sprintf("https://mangadex.org/titles/2/%d", index))
	}
	return urls, nil
}

func ScrapeMangaDexListing(ctx context.Context, workers int, pageLinks []string) <-chan DexScrapeResult {
	workchan := make(chan string)
	resultChan := make(chan DexScrapeResult)
	scraperManager := &MangaDexScraperManager{workerCount: workers, resultchan: &resultChan}
	scraperManager.dispatchWorkers(ctx, &workchan, &resultChan)
	go scraperManager.waitForWorkersDeath()
	go func() {
		for _, url := range pageLinks {
			workchan <- url
		}
	}()
	return resultChan
}

func cacheMangaListToFS(filePath string, mangas *[]scraper.Manga) error {
	mangabytes, err := json.Marshal(mangas)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, mangabytes, 0777)
	if err != nil {
		return err
	}
	return nil
}

func readMangaListFromFSCache(filePath string) (*[]scraper.Manga, error) {
	mangabytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	mangas := []scraper.Manga{}
	err = json.Unmarshal(mangabytes, &mangas)
	if err != nil {
		return nil, err
	}
	return &mangas, nil
}

func getMangaDexMangaList(useCache bool, cachePath string) ([]scraper.Manga, error) {
	if useCache {
		mangas, err := readMangaListFromFSCache(cachePath)
		if err == nil && len(*mangas) > 0 {
			return *mangas, nil
		}
		log.Printf("couldn't open cache at %s: %s\n", cachePath, err)
	}
	log.Println("trying to get the full mangadex listing...this will take a while")

	pageLinks, err := GetMangaDexListingPageLinks()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	resultChan := ScrapeMangaDexListing(ctx, 10, pageLinks)

	mangas := []scraper.Manga{}
	for i := len(pageLinks); i > 0; i-- {
		result := <-resultChan
		if result.Err == nil {
			mangas = append(mangas, result.Mangas...)
		}
	}

	return mangas, cacheMangaListToFS(cachePath, &mangas)
}

func searchMangaList(useCache bool, cachePath string, mangaName string) ([]scraper.Manga, error) {
	mangaList, err := getMangaDexMangaList(useCache, cachePath)
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

type respMangaInfo struct {
	Title string `json:"title"`
}

type respChapterInfo struct {
	Volume  string   `json:"volume"`
	Chapter string   `json:"chapter"`
	Title   string   `json:"title"`
	Lang    string   `json:"lang_code"`
	Hash    string   `json:"hash,omitEmpty"`
	Server  string   `json:"server,omitEmpty"`
	Pages   []string `json:"page_array,omitEmpty"`
}

func getMangaChapterList(mangaID string) ([]scraper.Chapter, error) {
	resp, err := scraper.MakeRequest(fmt.Sprintf("%s/manga/%s", MANGADEX_API_URL, mangaID))
	if err != nil {
		return []scraper.Chapter{}, err
	}

	respMap := make(map[string]*json.RawMessage)
	bodyByte, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyByte, &respMap)
	if err != nil {
		return []scraper.Chapter{}, err
	}

	mangaInfo := respMangaInfo{}
	err = json.Unmarshal(*respMap["manga"], &mangaInfo)
	if err != nil {
		return []scraper.Chapter{}, err
	}

	chMap := make(map[string]respChapterInfo)
	err = json.Unmarshal(*respMap["chapter"], &chMap)
	if err != nil {
		return []scraper.Chapter{}, err
	}

	chapters := []scraper.Chapter{}
	for k, v := range chMap {
		// no support for other langs for now
		if v.Lang == "gb" {
			chapters = append(chapters, scraper.Chapter{
				ID:           v.Chapter,
				ChapterTitle: v.Title,
				URL:          fmt.Sprintf("%s/chapter/%s", MANGADEX_API_URL, k),
				SourceName:   SOURCE_NAME,
				MangaName:    mangaInfo.Title,
			})
		}
	}
	return chapters, nil
}

func getChapterInfo(url string) (*respChapterInfo, error) {
	resp, err := scraper.MakeRequest(url)
	if err != nil {
		return nil, err
	}
	respChap := respChapterInfo{}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBytes, &respChap)
	if err != nil {
		return nil, err
	}
	return &respChap, nil
}

func getMangaChapter(mangaID, chapterID string) (scraper.Chapter, error) {
	chapterList, err := getMangaChapterList(mangaID)
	if err != nil {
		return scraper.Chapter{}, err
	}

	for _, chapter := range chapterList {
		// lol ok
		chInputAsFloat, _ := strconv.ParseFloat(chapterID, 32)
		chAsFloat, _ := strconv.ParseFloat(chapter.ID, 32)
		if chAsFloat == chInputAsFloat {
			chapterInfo, err := getChapterInfo(chapter.URL)
			if err != nil {
				return scraper.Chapter{}, err
			}
			chapterPages := []scraper.ChapterPage{}
			for i, page := range chapterInfo.Pages {
				imgURL := fmt.Sprintf("%s/%s/%s", chapterInfo.Server, chapterInfo.Hash, page)
				chapterPages = append(chapterPages, scraper.ChapterPage{Page: i, Url: imgURL})
			}
			chapter.ChapterPages = chapterPages
			return chapter, nil
		}
	}
	return scraper.Chapter{}, errors.New("chapter not found")
}

func (*MangaDex) Name() string {
	return SOURCE_NAME
}

func (dex *MangaDex) ListMangaDirectory() ([]scraper.Manga, error) {
	return getMangaDexMangaList(true, dex.CachePath)
}

func (dex *MangaDex) Search(mangaName string) ([]scraper.Manga, error) {
	return searchMangaList(true, dex.CachePath, mangaName)
}

func (*MangaDex) ListMangaChapters(mangaID string) ([]scraper.Chapter, error) {
	return getMangaChapterList(mangaID)
}

func (*MangaDex) GetChapter(mangaID, chapterID string) (scraper.Chapter, error) {
	return getMangaChapter(mangaID, chapterID)
}
