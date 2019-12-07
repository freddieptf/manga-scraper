package cli

import (
	"context"
	"fmt"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type downloadJob struct {
	chapter *scraper.Chapter
	archive bool
}

type jobCompletionStatus struct {
	chapter          *scraper.Chapter
	storagePath      string
	downloadComplete bool
	err              error
}

type downloadWorker struct {
	id         int
	jobChan    <-chan downloadJob
	resultChan *chan jobCompletionStatus
	ctx        context.Context
	notifyDone func()
}

type Downloader struct {
	queueSize int
	mux       sync.Mutex

	// initialized if we want to get notified of queue size changes
	queueSizeChangeChan *chan int

	maxParallelDownloads int
	jobChan              chan downloadJob
	jobResultChan        chan jobCompletionStatus
	ctx                  context.Context
}

func initDownloader(ctx context.Context, maxParallelDownloads int) *Downloader {
	jobChan := make(chan downloadJob)
	resultChan := make(chan jobCompletionStatus, maxParallelDownloads)
	return &Downloader{
		ctx:                  ctx,
		maxParallelDownloads: maxParallelDownloads,
		jobChan:              jobChan,
		jobResultChan:        resultChan,
	}
}

func (downloader *Downloader) newWorker(id int) downloadWorker {
	return downloadWorker{
		id:         id,
		jobChan:    downloader.jobChan,
		resultChan: &downloader.jobResultChan,
		ctx:        downloader.ctx,
		notifyDone: downloader.decQueueCounter,
	}
}

func (downloader *Downloader) incQueueCounter() {
	downloader.mux.Lock()
	downloader.queueSize++
	downloader.mux.Unlock()
}

func (downloader *Downloader) decQueueCounter() {
	downloader.mux.Lock()
	downloader.queueSize--
	downloader.mux.Unlock()
}

func (downloader *Downloader) getQueueCounter() int {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()
	return downloader.queueSize
}

func (downloader *Downloader) queueDownload(chapter *scraper.Chapter, archive bool) {
	downloader.jobChan <- downloadJob{chapter: chapter, archive: archive}
	downloader.incQueueCounter()
}

func (downloader *Downloader) initWorkers() {
	for i := 0; i < downloader.maxParallelDownloads; i++ {
		chDownloader := downloader.newWorker(i)
		go chDownloader.waitForWork()
	}
	go func() {
		for status := range downloader.jobResultChan {
			if *downloader.queueSizeChangeChan != nil {
				*downloader.queueSizeChangeChan <- downloader.getQueueCounter()
			}
			if status.downloadComplete {
				fmt.Printf("download successful: %s\n", status.storagePath)
				if status.err != nil {
					log.Printf("error: %s %s, %v\n", status.chapter.MangaName, status.chapter.ChapterTitle, status.err)
				}
			} else {
				fmt.Printf("download unsuccessful: %s %s, %v\n", status.chapter.MangaName, status.chapter.ChapterTitle, status.err)
				log.Printf("%v\n", status.err)
			}
		}
	}()
}

func (downloader *Downloader) waitTillQueueEmptyAndExit() {
	queueChan := make(chan int)
	downloader.queueSizeChangeChan = &queueChan
	if downloader.getQueueCounter() == 0 {
		fmt.Println("Downloads Done!")
		return
	}
	for size := range *downloader.queueSizeChangeChan {
		if size == 0 {
			close(downloader.jobResultChan)
			fmt.Println("Downloads Done!")
			return
		}
	}
}

func (worker *downloadWorker) waitForWork() {
	for {
		select {
		case job := <-worker.jobChan:
			completionStatus := jobCompletionStatus{downloadComplete: true, chapter: job.chapter}
			path, err := downloadChapter(job.chapter)
			if err != nil {
				completionStatus.downloadComplete = false
				completionStatus.err = err
			} else {
				if job.archive {
					err = cbzify(path)
					if err != nil {
						completionStatus.err = err
					}
				}
			}
			completionStatus.storagePath = path
			*worker.resultChan <- completionStatus
			worker.notifyDone()

		case <-worker.ctx.Done():
			return
		}
	}
}

// make the manga source dir under the apps root dir, i.e Manga in the home dir
// returns the path of the dir created
func createMangaSourceDir(sourceName string) (string, error) {
	sourceDir := filepath.Join(os.Getenv("HOME"), "Manga", sourceName)
	err := os.MkdirAll(sourceDir, 0777)
	if err != nil {
		return "", err
	}
	return sourceDir, nil
}

// create a chapter dir in the provided manga source dir path, return the its path
func createChapterDir(sourceDirPath *string, chapter *scraper.Chapter) (string, error) {
	var chapterPath string
	if chapter.VolumeTitle == "" {
		chapterPath = filepath.Join(*sourceDirPath, chapter.MangaName, fmt.Sprintf("%s - %s", chapter.ID, chapter.ChapterTitle))
	} else {
		chapterPath = filepath.Join(*sourceDirPath, chapter.MangaName, chapter.VolumeTitle, fmt.Sprintf("%s - %s", chapter.ID, chapter.ChapterTitle))
	}

	err := os.MkdirAll(chapterPath, 0777)
	if err != nil {
		return "", fmt.Errorf("Couldn't make %s directory: err %v", chapterPath, err)
	}

	return chapterPath, nil
}

// get image given the url, return a byte array to do whatever we want with it
func getImage(url string) (*[]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func saveImageToDisk(path string, img *[]byte) error {
	return ioutil.WriteFile(path, *img, 0655)
}

func downloadChapter(chapter *scraper.Chapter) (string, error) {
	sourceDirPath, err := createMangaSourceDir(chapter.SourceName)
	if err != nil {
		log.Fatalf("couldn't make source %s library dir: err %v", chapter.SourceName, err)
	}

	chapterPath, err := createChapterDir(&sourceDirPath, chapter)
	if err != nil {
		log.Fatalf("could not create chapterDir %s\n", err)
	}

	fmt.Printf("downloading: %s %s\n", chapter.MangaName, chapter.ChapterTitle)

	for _, chapterPage := range chapter.ChapterPages {
		pagePath := filepath.Join(chapterPath, strconv.Itoa(chapterPage.Page)+".jpg")
		img, err := getImage(chapterPage.Url)
		if err != nil {
			log.Printf("%v occured while getting %s\n", err, chapterPage.Url)
		} else {
			err = saveImageToDisk(pagePath, img)
			if err != nil {
				log.Printf("%v, couldn't save img to %s\n", err, pagePath)
			}
		}
	}
	return chapterPath, nil
}
