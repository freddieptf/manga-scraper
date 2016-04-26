package commons

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

type ImgItem struct {
	URL string
	ID  int
}

func DownloadImg(page int, url, path string) error {
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
