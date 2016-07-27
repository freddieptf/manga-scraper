package msources

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/go-playground/pool.v1"

	"github.com/PuerkitoBio/goquery"
)

const mangareaderURL = "http://www.mangareader.net"

type searchResult struct {
	manga, mangaID string
}

//GetFromReader get manga chapters from mangareader
func (d *MangaDownload) GetFromReader(n int) {
	results, err := searchReader(d.MangaName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Id \t Manga\n")
	for i, m := range results {
		fmt.Printf("%d \t %s\n", i, m.manga)
	}

	//get the correct id from the user incase of multiple match results
	myScanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter the id of the correct manga: ")
	var id int
scanDem:
	for myScanner.Scan() {
		id, err = strconv.Atoi(myScanner.Text())
		if err != nil {
			fmt.Printf("Enter a valid Id, please: ")
			goto scanDem
		}
		break
	}

	//get the matching id
	match, exists := results[id]
	if !exists {
		fmt.Printf("Insert one of the Ids in the results, please: ")
		goto scanDem
	}
	urlPath := match.mangaID
	*d.MangaName = match.manga

	p := pool.NewPool(n, len(*d.Args)) //goroutine pool

	fn := func(job *pool.Job) { //job
		e := job.Params()[0].(*chapterDownload).getChapterFromReader()
		if e != nil {
			fmt.Printf("Download Failed: %v chapter %v (%v)\n",
				job.Params()[0].(*chapterDownload).manga, job.Params()[0].(*chapterDownload).chapter, e)
			return
		}
		job.Return(job.Params()[0].(*chapterDownload).manga + " chapter " + job.Params()[0].(*chapterDownload).chapter)
	}

	for _, chapter := range *d.Args {
		p.Queue(fn, &chapterDownload{
			url:     urlPath, //we actually pass the manga_id from the path here and build the url later in getChapterFromReader
			manga:   *d.MangaName,
			chapter: strconv.Itoa(chapter),
		}) //queue the jobs
	}

	for result := range p.Results() { //receive results
		err, ok := result.(*pool.ErrRecovery)
		if ok { // there was some sort of panic that
			log.Println(err) // was recovered, in this scenario
			return
		}
		res := result.(string)
		fmt.Println("Download Successful: ", res)
	}

}

//@TODO If a chapter doesn't exist onsite, return an error.
func (c *chapterDownload) getChapterFromReader() error {
	var (
		urls    []string
		imgUrls []imgItem
		doc     *goquery.Document
		err     error
		baseURL = "http://www.mangareader.net"
	)

	doc, err = goquery.NewDocument(baseURL + c.url + "/" + c.chapter)
	if err != nil {
		return err
	}

	//get the manga page urls
	doc.Find("div#selectpage > select#pageMenu > option").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("value")
		url = baseURL + url
		urls = append(urls, url)
	})

	//go off and get the chapter title from the chapter listings
	titleChan := make(chan string)
	go func() {
		var title string
		d, e := goquery.NewDocument(baseURL + c.url)
		if e != nil {
			log.Println(e)
			titleChan <- title
			return
		}
		d.Find("div#chapterlist > table#listing td").Has("div.chico_manga").EachWithBreak(func(i int, s *goquery.Selection) bool {
			link, _ := s.Find("a").Attr("href")
			ps := strings.Split(link, "/")
			if ps[2] == c.chapter {
				title = strings.Split(s.Text(), " : ")[1]
				return false
			}
			return true
		})
		titleChan <- title
	}()

	//scrape the manga pages for the image urls
	fmt.Printf("%v %v: Getting the chapter image urls\n", c.manga, c.chapter)

	imgItemChan := make(chan imgItem)
	for i, url := range urls {
		go func(i int, url string) {
			doc, err = goquery.NewDocument(url)
			if err != nil {
				log.Println(err)
				return
			}
			imgURL, _ := doc.Find("div#imgholder img").Attr("src")
			imgItemChan <- imgItem{ID: i, URL: imgURL}
		}(i, url)
	}

	for i := 0; i < len(urls); i++ {
		imgUrls = append(imgUrls, <-imgItemChan)
	}
	chapterTitle := <-titleChan
	chapterPath := filepath.Join(os.Getenv("HOME"), "Manga", "MangaReader", c.manga, c.chapter+": "+chapterTitle)
	err = os.MkdirAll(chapterPath, 0777)
	if err != nil {
		log.Println("Couldn't make directory ", chapterPath)
		return err
	}
	fmt.Printf("Downloading %s %s to %v \n", c.manga, c.chapter, chapterPath)
	ch := make(chan error)
	for _, item := range imgUrls {
		go func(item imgItem) {
			err = item.downloadImage(chapterPath)
			if err != nil {
				ch <- err
			}
			ch <- nil
		}(item)
	}
	for range imgUrls {
		err := <-ch
		if err != nil {
			os.RemoveAll(chapterPath)
			return err
		}
	}

	err = cbzify(chapterPath)
	if err != nil {
		fmt.Printf("Couldn't make chapter cbz: %v\n", err)
	}

	return nil
}

func searchReader(manga *string) (map[int]searchResult, error) {
	doc, err := goquery.NewDocument(mangareaderURL + "/alphabetical")
	if err != nil {
		return nil, err
	}
	var results = make(map[int]searchResult)
	//find possible matches in the site's manga list for the mangaName provided;
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(*manga)) {
			mid, _ := s.Attr("href")
			results[i] = searchResult{s.Text(), mid}
		}
	})

	if len(results) <= 0 {
		return nil, errors.New("found Zero results. Exiting")
	}

	return results, nil
}
