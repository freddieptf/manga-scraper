package msources

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/go-playground/pool.v1"

	"github.com/PuerkitoBio/goquery"
)

func (d *MangaDownload) GetFromReader(n int) {
	doc, err := goquery.NewDocument("http://www.mangareader.net/alphabetical")
	if err != nil {
		log.Fatal(err)
	}

	var matches = make(map[int]string)
	var matchesNames = make(map[int]string)
	//find possible matches in the site's manga list for the mangaName provided;
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(*d.MangaName)) {
			matches[i], _ = s.Attr("href")
			matchesNames[i] = s.Text()
		}
	})

	if len(matches) <= 0 {
		log.Fatal(*d.MangaName + " could not be found")
	}

	fmt.Printf("Id \t Manga\n")
	for i, m := range matches {
		fmt.Printf("%d \t %s\n", i, m)
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
	urlPath, exists := matches[id]
	if !exists {
		fmt.Printf("Insert one of the Ids in the results, please: ")
		goto scanDem
	}
	*d.MangaName = matchesNames[id]

	p := pool.NewPool(n, len(*d.Chapters)) //goroutine pool

	fn := func(job *pool.Job) { //job
		ch, e := getChapterFromReader(job.Params()[0].(string), job.Params()[1].(string), job.Params()[2].(string))
		if e != nil {
			fmt.Printf("Download Failed: %v chapter %v (%v)\n", job.Params()[1].(string), job.Params()[2].(string), e)
			return
		}
		job.Return(job.Params()[1].(string) + " " + ch)
	}

	for _, chapter := range *d.Chapters {
		p.Queue(fn, urlPath, *d.MangaName, strconv.Itoa(chapter)) //queue the jobs
	}

	for result := range p.Results() { //receive results
		err, ok := result.(*pool.ErrRecovery)
		if ok { // there was some sort of panic that
			fmt.Println(err) // was recovered, in this scenario
			return
		}
		res := result.(string)
		fmt.Println("Download Successful: ", res)
	}

}

//@TODO If a chapter doesn't exist onsite, return an error.
func getChapterFromReader(mangaPath, mangaName, chapter string) (string, error) {
	var (
		urls    []string
		imgUrls []imgItem
		doc     *goquery.Document
		err     error
		baseURL = "http://www.mangareader.net"
	)

	doc, err = goquery.NewDocument(baseURL + mangaPath + "/" + chapter)
	if err != nil {
		return chapter, err
	}

	//get the manga page urls
	fmt.Printf("%v %v: Getting the Manga page urls\n", mangaName, chapter)
	doc.Find("div#selectpage > select#pageMenu > option").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("value")
		url = baseURL + url
		urls = append(urls, url)
	})

	//go off and get the chapter title from the chapter listings
	titleChan := make(chan string)
	go func() {
		var title string
		d, e := goquery.NewDocument(baseURL + mangaPath)
		if e != nil {
			log.Println(e)
			titleChan <- title
			return
		}
		d.Find("div#chapterlist > table#listing td").Has("div.chico_manga").EachWithBreak(func(i int, s *goquery.Selection) bool {
			link, _ := s.Find("a").Attr("href")
			ps := strings.Split(link, "/")
			if ps[2] == chapter {
				title = strings.Split(s.Text(), " : ")[1]
				return false
			}
			return true
		})
		titleChan <- title
	}()

	//scrape the manga pages for the image urls
	fmt.Printf("%v %v: Getting the chapter image urls\n", mangaName, chapter)

	imgItemChan := make(chan imgItem)
	for i, url := range urls {
		go func(i int, url string) {
			doc, err = goquery.NewDocument(url)
			if err != nil {
				log.Println(err)
			}
			imgURL, _ := doc.Find("div#imgholder img").Attr("src")
			imgItemChan <- imgItem{ID: i, URL: imgURL}
		}(i, url)
	}

	for i := 0; i < len(urls); i++ {
		imgUrls = append(imgUrls, <-imgItemChan)
	}

	chapterPath := filepath.Join(os.Getenv("HOME"), "Manga", "MangaReader", mangaName, chapter+": "+<-titleChan)
	err = os.MkdirAll(chapterPath, 0777)
	if err != nil {
		log.Fatal("Couldn't make directory ", err)
	}
	fmt.Printf("Downloading %s %s to %v: \n", mangaName, chapter, chapterPath)
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
			return chapter, err
		}
	}

	return chapter, nil
}
