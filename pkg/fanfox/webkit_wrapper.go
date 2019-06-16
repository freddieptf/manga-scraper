package fanfox

import (
	"errors"

	"github.com/freddieptf/go-webkit2/webkit2"
	"github.com/gotk3/gotk3/glib"
	"github.com/sqs/gojs"
)

// shamelessly copied from https://github.com/sourcegraph/webloop/blob/master/webloop.go
// don't want a dependency on webloop, just wanna use webkit

// ErrLoadFailed indicates that the View failed to load the requested resource.
var ErrLoadFailed = errors.New("load failed")

// Context stores common settings for a group of Views.
type context struct{}

// New creates a new Context.
func New() *context {
	return &context{}
}

// NewView creates a new View in the context.
func (c *context) NewView() *webviewWrapper {
	view := make(chan *webviewWrapper, 1)
	glib.IdleAdd(func() bool {
		webView := webkit2.NewWebView()
		settings := webView.Settings()
		settings.SetEnableWriteConsoleMessagesToStdout(true)
		v := &webviewWrapper{WebView: webView}
		loadChangedHandler, _ := webView.Connect("load-changed", func(_ *glib.Object, loadEvent webkit2.LoadEvent) {
			switch loadEvent {
			case webkit2.LoadFinished:
				// If we're here, then the load must not have failed, because
				// otherwise we would've disconnected this handler in the
				// load-failed signal handler.
				v.load <- struct{}{}
			}
		})
		webView.Connect("load-failed", func() {
			v.lastLoadErr = ErrLoadFailed
			webView.HandlerDisconnect(loadChangedHandler)
		})
		view <- v
		return false
	})
	return <-view
}

// View represents a WebKit view that can load resources at a given URL and
// query information about them.
type webviewWrapper struct {
	*webkit2.WebView

	load        chan struct{}
	lastLoadErr error

	destroyed bool
}

// Open starts loading the resource at the specified URL.
func (v *webviewWrapper) Open(url string) {
	v.load = make(chan struct{}, 1)
	v.lastLoadErr = nil
	glib.IdleAdd(func() bool {
		if !v.destroyed {
			v.WebView.LoadURI(url)
		}
		return false
	})
}

func (v *webviewWrapper) Load(content, baseUrl string) {
	v.load = make(chan struct{}, 1)
	v.lastLoadErr = nil
	glib.IdleAdd(func() bool {
		if !v.destroyed {
			v.WebView.LoadHTML(content, baseUrl)
		}
		return false
	})
}

// Wait waits for the current page to finish loading.
func (v *webviewWrapper) Wait() error {
	<-v.load
	return v.lastLoadErr
}

// EvaluateJavaScript runs the JavaScript in script in the view's context and
// returns the script's result as a Go value.
func (v *webviewWrapper) EvaluateJavaScript(script string) (result interface{}, err error) {
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	glib.IdleAdd(func() bool {
		v.WebView.RunJavaScript(script, func(result *gojs.Value, err error) {
			glib.IdleAdd(func() bool {
				if err == nil {
					goval, err := result.GoValue()
					if err != nil {
						errChan <- err
						return false
					}
					resultChan <- goval
				} else {
					errChan <- err
				}
				return false
			})
		})
		return false
	})

	select {
	case result = <-resultChan:
		return result, nil
	case err = <-errChan:
		return nil, err
	}
}

// Close closes the view and releases associated resources. Ensure that Close is
// called after all other pending operations on View have returned, or they may
// hang indefinitely.
func (v *webviewWrapper) Close() {
	// TODO(sqs): remove all of the source funcs we added via IdleAdd, etc.,
	// using g_source_remove, to fix "assertion
	// 'WEBKIT_IS_WEB_VIEW(webView) failed" messages.
	v.destroyed = true
	v.Destroy()
}
