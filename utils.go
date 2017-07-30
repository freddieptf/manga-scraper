package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func GetRange(vals *[]string) *[]int {
	var x, y int
	var err error
	var chapters []int
	for _, val := range *vals {
		if strings.Contains(val, "-") {
			chs := strings.Split(val, "-")
			x, err = strconv.Atoi(chs[0])
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			y, err = strconv.Atoi(chs[1])
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			for i := x; i <= y; i++ {
				chapters = append(chapters, i)
			}
		} else {
			x, err = strconv.Atoi(val)
			if err != nil {
				log.Printf("%v could not be converted to a chapter.\n", val)
				log.Fatal(err)
			}
			chapters = append(chapters, x)
		}
	}

	fmt.Printf("%v\n", chapters)

	return &chapters
}

func getMatchFromSearchResults(results map[int]searchResult) searchResult {
	fmt.Printf("Id \t Manga\n")
	for i, m := range results {
		fmt.Printf("%d \t %s\n", i, m.manga)
	}

	myScanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter the id of the correct manga: ")
	var id int
	var err error
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
	return match
}
