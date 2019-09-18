package scraper

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	transport *http.Transport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: 40 * time.Second,
	}
	client    *http.Client = &http.Client{Transport: transport, Timeout: 1 * time.Minute}
	userAgent              = "Mozilla/5.0 (X11; Linux x86_64; rv:66.0) Gecko/20100101 Firefox/66.0"
)

func MakeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return client.Do(req)
}

func MakeDocRequest(url string) (*goquery.Document, error) {
	resp, err := MakeRequest(url)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromResponse(resp)
}
