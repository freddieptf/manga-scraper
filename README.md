[![pipeline status](https://gitlab.com/freddieptf/manga-scraper/badges/master/pipeline.svg)](https://gitlab.com/freddieptf/manga-scraper/commits/master)

```
go get github.com/freddieptf/manga-scraper
```

## Usage

The general usage format is

	./manga-scraper --{source} "manga name" chapters

valid sources are `mf` for mangafox and `mr` for mangareader

*If a source isn't provided, mangareader is used as the default*

##### Download a single chapter.

  	./manga-scraper --{source} "manga name" chapter

*example*

	./manga-scraper --mf "dokgo" 2


##### Download chapters over a certain range

	./manga-scraper --{source} "manga name" start-stop

*example*

	./manga-scraper --mf "dokgo" 3-24

The default maximum number of active concurrent downloads is 1. If the downloads are >1, they are queue'd. This can be changed by defining your own like: 

	./manga-scraper --mr --n=2 "dokgo" 2-30

This changes the max to 2 and queues the rest until a download slot is free. Keep these value as low as possible to avoid hammering the sites and reduce the chance of errors and failures, If you go any higher, the downloads start randomly failing.

##### All together now

	./manga-scraper --mr --n=6 "dokgo" 2 10-24 27 34 36-46

##### Downloading volumes (only on mangafox)
This will download volume 1 to 5 of dokgo

	./manga-scraper --mf --vlm "dokgo" 1-5




#### Need some quick Help?

	./manga-scraper --help




Downloads are kept in the users home directory in the format `Manga/{source}/{mangaName}/{chapter}`

### To be added/fixed:
 - [x] Add support for unordered chapter downloads, `1 2 45 22 56`
 - [ ] Fix failed downloads
 - [ ] Support updating all the mangas in the manga directory to the latest chapters
 - [x] Support volume download(mangafox)
 - [x] Threadpoolish implementation for downloads
 - [x] Download into `.cbz` format
