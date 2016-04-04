package source

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Get : get chapters
func Get(start, stop int, mangaName string) {
	// l, _ := os.Open("list.html")
	doc, err := goquery.NewDocument("http://www.mangareader.net/alphabetical")
	// defer l.Close()
	// doc, err := goquery.NewDocumentFromReader(l)
	if err != nil {
		log.Fatal(err)
	}

	var matches = make(map[int]string)
	//find possible matches in the site's manga list for the mangaName provided;
	doc.Find("ul.series_alpha > li > a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(mangaName)) {
			matches[i], _ = s.Attr("href")
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

	if stop == -1 { //for when we're downloading a single chapter
		stop = start
	}

	ch := make(chan int)
	for i := start; i <= stop; i++ {
		go func(urlPath, mangaName string, i int) {
			Chapter, theError := getDemChapters(urlPath, mangaName, strconv.Itoa(i))
			if theError != nil {
				fmt.Printf("Download Failed: %v chapter %v (%v)\n", mangaName, Chapter, theError)
			} else {
				fmt.Printf("Download done: %v chapter %v\n", mangaName, Chapter)
			}
			ch <- i
		}(urlPath, mangaName, i)
	}
	for i := start; i <= stop; i++ {
		<-ch
	}

}

func getDemChapters(urlPath, mangaName, chapter string) (string, error) {
	var (
		urls, imgurls []string
		doc           *goquery.Document
		err           error
		baseUrl       = "http://www.mangareader.net"
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

	//scrape the manga pages for the image urls
	fmt.Printf("%v %v: Getting the chapter image urls\n", mangaName, chapter)
	for i := 0; i < len(urls); i++ {
		doc, err = goquery.NewDocument(urls[i])
		if err != nil {
			return chapter, err
		}
		imgURL, _ := doc.Find("div#imgholder img").Attr("src")
		imgurls = append(imgurls, imgURL)
	}

	chapterPath := filepath.Join("Manga", mangaName, chapter)
	err = os.MkdirAll(chapterPath, 0655) //0644 gadamit
	if err != nil {
		log.Fatal("You might want to run it as SU. Couldn't make directory ", err)
	}
	fmt.Printf("Downloading %s %s: \n", mangaName, chapter)
	ch := make(chan error)
	for i, url := range imgurls {
		go func(i int, url string) {
			err = downloadImg(i, url, mangaName, chapter, chapterPath)
			if err != nil {
				ch <- err
			}
			ch <- nil
		}(i, url)
	}
	for range imgurls {
		err := <-ch
		if err != nil {
			os.RemoveAll(chapterPath)
			return chapter, err
		}
	}

	return chapter, nil
}

func downloadImg(page int, url, mangaName, chapter, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	imgPath := filepath.Join(path, strconv.Itoa(page)+".jpg")
	err = ioutil.WriteFile(imgPath, body, 0655)
	if err != nil {
		return err
	}
	return nil
}
