package source

import (
	utils "GoManga/commons"
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Get : get chapters
func Get(chapters []int, mangaName string) {
	// l, _ := os.Open("list.html")
	doc, err := goquery.NewDocument("http://www.mangareader.net/alphabetical")
	// defer l.Close()
	// doc, err := goquery.NewDocumentFromReader(l)
	if err != nil {
		log.Fatal(err)
	}

	var matches = make(map[int]string)
	var matchesNames = make(map[int]string)
	//find possible matches in the site's manga list for the mangaName provided;
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(mangaName)) {
			matches[i], _ = s.Attr("href")
			matchesNames[i] = s.Text()
		}
	})

	if len(matches) <= 0 {
		log.Fatal(mangaName + " could not be found")
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
	mangaName = matchesNames[id]

	ch := make(chan int)
	for _, chapter := range chapters {
		go func(urlPath, mangaName string, chapter int) {
			Chapter, theError := getDemChapters(urlPath, mangaName, strconv.Itoa(chapter))
			if theError != nil {
				fmt.Printf("Download Failed: %v chapter %v (%v)\n", mangaName, Chapter, theError)
			} else {
				fmt.Printf("Download done: %v chapter %v\n", mangaName, Chapter)
			}
			ch <- chapter
		}(urlPath, mangaName, chapter)
	}
	for range chapters {
		<-ch
	}

}

func getDemChapters(urlPath, mangaName, chapter string) (string, error) {
	var (
		urls    []string
		imgUrls []utils.ImgItem
		doc     *goquery.Document
		err     error
		baseUrl = "http://www.mangareader.net"
	)

	doc, err = goquery.NewDocument(baseUrl + urlPath + "/" + chapter)
	if err != nil {
		return chapter, err
	}

	//get the manga page urls
	fmt.Printf("%v %v: Getting the Manga page urls\n", mangaName, chapter)
	doc.Find("div#selectpage > select#pageMenu > option").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("value")
		url = baseUrl + url
		urls = append(urls, url)
	})

	//go off and get the chapter title from the chapter listings
	titleChan := make(chan string)
	go func() {
		var title string
		d, e := goquery.NewDocument(baseUrl + urlPath)
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

	imgItemChan := make(chan utils.ImgItem)
	for i, url := range urls {
		go func(i int, url string) {
			doc, err = goquery.NewDocument(url)
			if err != nil {
				log.Println(err)
			}
			imgURL, _ := doc.Find("div#imgholder img").Attr("src")
			imgItemChan <- utils.ImgItem{ID: i, URL: imgURL}
		}(i, url)
	}

	for i := 0; i < len(urls); i++ {
		imgUrls = append(imgUrls, <-imgItemChan)
	}

	home := os.Getenv("HOME")
	chapterPath := filepath.Join(home, "Manga", "MangaReader", mangaName, chapter+": "+<-titleChan)
	err = os.MkdirAll(chapterPath, 0777)
	if err != nil {
		log.Fatal("Couldn't make directory ", err)
	}
	fmt.Printf("Downloading %s %s to %v: \n", mangaName, chapter, chapterPath)
	ch := make(chan error)
	for _, item := range imgUrls {
		go func(imgItem utils.ImgItem) {
			err = utils.DownloadImg(imgItem.ID, imgItem.URL, chapterPath)
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
