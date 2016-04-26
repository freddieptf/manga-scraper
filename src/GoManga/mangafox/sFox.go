package mangafox

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

const (
	MANGAFOX_URL string = "http://mangafox.com/"
)

func Get(start, stop int, mangaName string) {
	doc, err := goquery.NewDocument(MANGAFOX_URL + "manga/")
	if err != nil {
		log.Fatal(err)
		return
	}

	var matches = make(map[int]string)
	var matchesNames = make(map[int]string)
	doc.Find("div.manga_list li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(mangaName)) {
			matches[i], _ = s.Attr("href")
			matchesNames[i] = s.Text()
		}
	})

	if len(matches) <= 0 {
		log.Fatal(mangaName + " could not be found")
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
	mangaURL, exists := matches[id]
	if !exists {
		fmt.Printf("Insert one of the Ids in the results, please: ")
		goto scanDem
	}
	mangaName = matchesNames[id]

	if stop == -1 { //for when we're downloading a single chapter
		stop = start
	}

	ch := make(chan int)
	for i := start; i <= stop; i++ {
		go func(urlPath, mangaName string, i int) {
			Chapter, theError := getDemChapters(mangaURL, mangaName, strconv.Itoa(i))
			if theError != nil {
				fmt.Printf("Download Failed: %v chapter %v (%v)\n", mangaName, Chapter, theError)
			} else {
				fmt.Printf("Download done: %v chapter %v\n", mangaName, Chapter)
			}
			ch <- i
		}(mangaURL, mangaName, i)
	}
	for i := start; i <= stop; i++ {
		<-ch
	}

}

func getDemChapters(mangaURL, mangaName, chapter string) (string, error) {
	doc, err := goquery.NewDocument(mangaURL)
	if err != nil {
		return chapter, err
	}
	var page1 string
	name, _ := doc.Find("div.cover img").Attr("alt")
	var urls []string
	var imgUrls []utils.ImgItem

	doc.Find("ul.chlist li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		chId := strings.TrimPrefix(s.Find("a").Text(), name+" ")
		if chapter == chId {
			page1, _ = s.Find("a").Last().Attr("href")
			return false
		}
		return true
	})

	baseUrl := strings.TrimSuffix(page1, "1.html")
	doc, err = goquery.NewDocument(baseUrl)
	if err != nil {
		return chapter, err
	}

	titleChan := make(chan string)
	go func(doc *goquery.Document) {
		titleChan <- strings.Split(doc.Find("div#tip").Find("strong").First().Text(), ": ")[1]
	}(doc)

	doc.Find("div#top_center_bar select.m option").Each(func(i int, s *goquery.Selection) {
		url_id := s.Text()
		urls = append(urls, baseUrl+url_id+".html")
	})

	if len(urls) == 0 {
		return "OOPS. CAN'T GET DIS: " + chapter, nil
	}

	imgItemChan := make(chan utils.ImgItem)
	for i, url := range urls[:len(urls)-1] {
		go func(i int, url string) {
			doc, _ = goquery.NewDocument(url)
			imgURL, _ := doc.Find("div.read_img img").Attr("src")
			imgItemChan <- utils.ImgItem{URL: imgURL, ID: i}
		}(i, url)
	}

	for i := 0; i < len(urls)-1; i++ {
		imgUrls = append(imgUrls, <-imgItemChan)
	}

	home := os.Getenv("HOME")
	chapterPath := filepath.Join(home, "Manga", "MangaFox", mangaName, chapter+": "+<-titleChan)
	err = os.MkdirAll(chapterPath, 0655) //0644 gadamit
	if err != nil {
		log.Fatal("You might want to run it as SU. Couldn't make directory ", err)
	}
	fmt.Printf("Downloading %s %s to %v: \n", mangaName, chapter, chapterPath)
	ch := make(chan error)
	for _, item := range imgUrls {
		go func(item utils.ImgItem) {
			err = utils.DownloadImg(item.ID, item.URL, chapterPath)
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
