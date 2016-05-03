package msources

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	mangafoxURL string = "http://mangafox.com/"
)

func (d *MangaDownload) GetFromFox() {
	doc, err := goquery.NewDocument(mangafoxURL + "manga/")
	if err != nil {
		log.Fatal(err)
		return
	}

	var matches = make(map[int]string)
	var matchesNames = make(map[int]string)
	doc.Find("div.manga_list li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(*d.MangaName)) {
			matches[i], _ = s.Attr("href")
			matchesNames[i] = s.Text()
		}
	})

	if len(matches) <= 0 {
		log.Fatal(*d.MangaName + " could not be found")
	}

	fmt.Printf("Id \t Manga\n")
	for i, m := range matchesNames {
		fmt.Printf("%d \t %s\n", i, m)
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
	mangaURL, exists := matches[id] // mangafox has the manga url also in the catalogue so we use that
	if !exists {
		fmt.Printf("Insert one of the Ids in the results, please: ")
		goto scanDem
	}
	*d.MangaName = matchesNames[id]

	ch := make(chan int)
	for _, chapter := range *d.Chapters {
		go func(urlPath, mangaName string, chapter int) {
			Chapter, theError := getChapterFromFox(mangaURL, mangaName, strconv.Itoa(chapter))
			if theError != nil {
				fmt.Printf("Download Failed: %v chapter %v (%v)\n", mangaName, Chapter, theError)
			} else {
				fmt.Printf("Download done: %v chapter %v\n", mangaName, Chapter)
			}
			ch <- chapter
		}(mangaURL, *d.MangaName, chapter)
	}
	for range *d.Chapters {
		<-ch
	}

}

func getChapterFromFox(mangaURL, mangaName, chapter string) (string, error) {
	doc, err := goquery.NewDocument(mangaURL) //open the manga's page on mangafox
	if err != nil {
		return chapter, err
	}
	var page1 string
	name, _ := doc.Find("div.cover img").Attr("alt") //get the name of the manga for uniformity when creating it's download dir
	var urls []string
	var imgUrls []imgItem

	doc.Find("ul.chlist li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		chID := strings.TrimPrefix(s.Find("a").Text(), name+" ")
		if chapter == chID { // search for the matching chapter in the manga's chapter catalogue
			page1, _ = s.Find("a").Last().Attr("href")
			return false
		}
		return true
	})

	baseURL := strings.TrimSuffix(page1, "1.html")
	doc, err = goquery.NewDocument(baseURL) //get the chapter's page
	if err != nil {
		return chapter, err
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
		return "OOPS. CAN'T GET DIS: " + chapter, nil
	}

	imgItemChan := make(chan imgItem)
	for i, url := range urls[:len(urls)-1] {
		go func(i int, url string) {
			doc, _ = goquery.NewDocument(url)                     //open a chapter page
			imgURL, _ := doc.Find("div.read_img img").Attr("src") //get the image url
			imgItemChan <- imgItem{URL: imgURL, ID: i}            //send it
		}(i, url)
	}

	for i := 0; i < len(urls)-1; i++ {
		imgUrls = append(imgUrls, <-imgItemChan) //get dem image urls
	}

	chapterPath := filepath.Join(os.Getenv("HOME"), "Manga", "MangaFox", mangaName, chapter+": "+<-titleChan)
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
			os.RemoveAll(chapterPath) //bad but for now...
			return chapter, err
		}
	}

	return chapter, nil
}
