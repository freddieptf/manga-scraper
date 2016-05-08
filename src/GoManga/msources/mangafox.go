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

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/go-playground/pool.v1"
)

const (
	mangafoxURL string = "http://mangafox.com/"
)

//GetFromFox gets manga chapters from mangafox
func (d *MangaDownload) GetFromFox(n int) {
	doc, err := goquery.NewDocument(mangafoxURL + "manga/")
	if err != nil {
		log.Fatal(err)
		return
	}

	var results = make(map[int]searchResult)

	doc.Find("div.manga_list li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(*d.MangaName)) {
			mid, _ := s.Attr("href")
			results[i] = searchResult{s.Text(), mid}
		}
	})

	if len(results) <= 0 {
		log.Fatal(*d.MangaName + " could not be found")
	}

	fmt.Printf("Id \t Manga\n")
	for i, m := range results {
		fmt.Printf("%d \t %s\n", i, m.manga)
	}

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
	match, exists := results[id] // mangafox has the manga url also in the catalogue so we use that
	if !exists {
		fmt.Printf("Insert one of the Ids in the results, please: ")
		goto scanDem
	}
	mangaURL := match.mangaID
	*d.MangaName = match.manga

	p := pool.NewPool(n, len(*d.Chapters))

	fn := func(job *pool.Job) {
		e := job.Params()[0].(*chapterDownload).getChapterFromFox()
		if e != nil {
			fmt.Printf("Download Failed: %v chapter %v (%v)\n",
				job.Params()[0].(*chapterDownload).manga, job.Params()[0].(*chapterDownload).chapter, e)
			return
		}
		job.Return(job.Params()[0].(*chapterDownload).manga + " " + job.Params()[0].(*chapterDownload).chapter)
	}

	for _, chapter := range *d.Chapters {
		p.Queue(fn, &chapterDownload{
			url:     mangaURL,
			manga:   *d.MangaName,
			chapter: strconv.Itoa(chapter),
		})
	}

	for result := range p.Results() {
		err, ok := result.(*pool.ErrRecovery)
		if ok { // there was some sort of panic that
			fmt.Println(err) // was recovered, in this scenario
			return
		}
		res := result.(string)
		fmt.Println("Download Successful: ", res)
	}

}

//GetVolumeFromFox gets manga volumes from Mangafox
func (d *MangaDownload) GetVolumeFromFox(n int) {
	fmt.Println(d)
}

func (c *chapterDownload) getChapterFromFox() error {
	doc, err := goquery.NewDocument(c.url) //open the manga's page on mangafox
	if err != nil {
		return err
	}
	var page1 string
	name, _ := doc.Find("div.cover img").Attr("alt") //get the name of the manga for uniformity when creating it's download dir
	var urls []string
	var imgUrls []imgItem

	doc.Find("ul.chlist li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		chID := strings.TrimPrefix(s.Find("a").Text(), name+" ")
		if c.chapter == chID { // search for the matching chapter in the manga's chapter catalogue
			page1, _ = s.Find("a").Last().Attr("href")
			return false
		}
		return true
	})

	baseURL := strings.TrimSuffix(page1, "1.html")
	doc, err = goquery.NewDocument(baseURL) //get the chapter's page
	if err != nil {
		return err
	}

	titleChan := make(chan string)
	go func(doc *goquery.Document) { //goroutine to get the chapter's title
		titleChan <- strings.Split(doc.Find("div#tip").Find("strong").First().Text(), ": ")[1]
	}(doc)

	//get the num of chapter pages so we can build all the page urls
	doc.Find("div#top_center_bar select.m option").Each(func(i int, s *goquery.Selection) {
		urlID := s.Text()
		urls = append(urls, baseURL+urlID+".html")
	})

	if len(urls) == 0 {
		return errors.New("OOPS. CAN'T GET DIS: " + c.chapter)
	}

	fmt.Printf("%v %v: Getting the chapter image urls\n", c.manga, c.chapter)
	imgItemChan := make(chan imgItem)
	for i, url := range urls[:len(urls)-1] {
		go func(i int, url string) {
			doc, err = goquery.NewDocument(url) //open a chapter page
			if err != nil {
				log.Println(err)
				return
			}
			imgURL, _ := doc.Find("div.read_img img").Attr("src") //get the image url
			imgItemChan <- imgItem{URL: imgURL, ID: i}            //send it
		}(i, url)
	}

	for i := 0; i < len(urls)-1; i++ {
		imgUrls = append(imgUrls, <-imgItemChan) //get dem image urls
	}

	chapterPath := filepath.Join(os.Getenv("HOME"), "Manga", "MangaFox", c.manga, c.chapter+": "+<-titleChan)
	err = os.MkdirAll(chapterPath, 0777)
	if err != nil {
		log.Fatal("Couldn't make directory ", err)
	}
	fmt.Printf("Downloading %s %s to %v: \n", c.manga, c.chapter, chapterPath)
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
			os.RemoveAll(chapterPath) //bad but for now...
			return err
		}
	}

	err = cbzify(chapterPath)
	if err != nil {
		fmt.Printf("Couldn't make chapter cbz: %v", err)
	}

	return nil
}
