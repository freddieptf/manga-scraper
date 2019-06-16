package fanfox

import (
	"fmt"
	"github.com/freddieptf/go-webkit2/webkit2"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"golang.org/x/net/html"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func makeDocRequestWebKit(url string) (*goquery.Document, error) {
	runtime.LockOSThread()
	gtk.Init(nil)

	docChan := make(chan *goquery.Document, 1)
	errChan := make(chan error, 1)

	go func() {
		ctx := New()
		view := ctx.NewView()

		settings := view.Settings()
		settings.SetUserAgentWithApplicationDetails("Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:62.0)", "Firefox/62.0")

		defer view.Close()
		view.Open(url)
		err := view.Wait()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load URL: %s\n", err)
			docChan <- nil
			errChan <- err
			gtk.MainQuit()
		}

		doc, err := getDocFromWebView(view)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't get doc from webview, %v\n", err)
			docChan <- nil
			errChan <- err
			gtk.MainQuit()
		}

		err = processDoc(url, doc, view)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't process doc: %s\n", err)
			docChan <- nil
			errChan <- err
			gtk.MainQuit()
		}

		docChan <- doc
		errChan <- nil
		gtk.MainQuit()

	}()

	gtk.Main()
	runtime.UnlockOSThread()

	return <-docChan, <-errChan

}

func getDocFromWebView(view *webviewWrapper) (*goquery.Document, error) {
	res, err := view.EvaluateJavaScript("document.documentElement.innerHTML")
	if err != nil {
		return nil, err
	}
	node, err := html.Parse(strings.NewReader(res.(string)))
	if err != nil {
		return nil, err
	}
	doc := goquery.NewDocumentFromNode(node)
	return doc, nil
}

func processDoc(sourceUrl string, doc *goquery.Document, view *webviewWrapper) error {
	host, err := url.ParseRequestURI(sourceUrl)
	if err != nil {
		return err
	}

	foxURLURL, _ := url.ParseRequestURI(foxURL)
	loadFinished := make(chan interface{}, 1)
	switch host.Host {
	case foxURLURL.Host:
		if doc.Find(".detail-block-content").Length() > 0 {

			view.WebView.Connect("load-changed", func(_ *glib.Object, i int) {
				loadEvent := webkit2.LoadEvent(i)
				switch loadEvent {
				case webkit2.LoadFinished:
					fmt.Println("load finished")
					loadFinished <- 1
				}
			})
			_, err := view.EvaluateJavaScript("if(typeof $('#checkAdult')[0] !== 'undefined'){ $('#checkAdult')[0].click();}location.reload();")
			if err != nil {
				return err
			}

			<-loadFinished

			d, err := getDocFromWebView(view)
			*doc = *d

		}
	}
	return nil
}
