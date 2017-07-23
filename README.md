# GoManga
A little program to scrape/download manga off [mangareader](http://www.mangareader.net/) and [mangafox](http://www.mangafox.me/).
*Other sources might be added..eventually*

```
go get github.com/freddieptf/manga-scraper
```

## Usage

The general usage format is

	./GoManga --{source} "manga name" chapters

valid sources are `mf` for mangafox and `mr` for mangareader

*If a source isn't provided, mangareader is used as the default*

##### Download a single chapter.

  	./GoManga --{source} "manga name" chapter

*example*

	./GoManga --mf "dokgo" 2


##### Download chapters over a certain range

	./GoManga --{source} "manga name" start-stop

*example*

	./GoManga --mf "dokgo" 3-24

The default maximum number of active concurrent downloads is 5 for mangareader and 1 for mangafox. If the downloads are >5 or >1 respectively, they are queue'd. This can be changed by defining your own like: 

	./GoManga --mr --n=2 "dokgo" 2-30

This changes the max to 2 and queues the rest until a download slot is free. Keep these value as low as possible to avoid hammering the sites and reduce the chance of errors and failures. The default value for mangafox is 1 since if you go any higher, the downloads start randomly failing.

##### All together now

	./GoManga --mr --n=6 "dokgo" 2 10-24 27 34 36-46

##### Downloading volumes (only on mangafox)
This will download volume 1 to 5 of dokgo

	./GoManga --mf --vlm "dokgo" 1-5




#### Need some quick Help?

	./GoManga --help




Downloads are kept in the users home directory in the format `Manga/{source}/{mangaName}/{chapter}`

### To be added/fixed:
 - [x] Add support for unordered chapter downloads, `1 2 45 22 56`
 - [ ] Fix failed downloads
 - [ ] Support updating all the mangas in the manga directory to the latest chapters
 - [x] Support volume download(mangafox)
 - [x] Threadpoolish implementation for downloads
 - [x] Download into `.cbz` format
