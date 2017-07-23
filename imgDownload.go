package main

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

type imgItem struct {
	URL string
	ID  int
}

func (item *imgItem) downloadImage(path string) error {
	response, err := http.Get(item.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	imgPath := filepath.Join(path, strconv.Itoa(item.ID)+".jpg")
	err = ioutil.WriteFile(imgPath, body, 0655)
	if err != nil {
		return err
	}
	return nil
}
