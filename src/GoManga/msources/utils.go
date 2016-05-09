package msources

import (
	"fmt"
	"log"
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
