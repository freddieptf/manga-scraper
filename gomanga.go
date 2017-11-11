package main

import (
	"fmt"

	cli "github.com/freddieptf/manga-scraper/cli"
)

func main() {
	fmt.Println("Start")
	cli.CliParse()
}
