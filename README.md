# GoManga
A little program to scrape/download manga off [mangareader] (http://www.mangareader.net/) and [mangafox] (http://www.mangafox.me/).
*Other sources might be added..eventually*

## Usage
Grab the executable in the bin folder and run it in the terminal..or compile the source if you Golang

The general usage format is

	./GoManga -source "manga name" chapters

valid sources are `-mf` for mangafox and `-mr` for mangareader

*If a source isn't provided, mangareader is used as the default*

##### Download a single chapter.

  	./GoManga -source "manga name" chapter

*example*

	./GoManga -mf "dokgo" 2


##### Download chapters over a certain range
*(advisable to keep the ranges short, <=50...for now.)*

    ./GoManga -source "manga name" start-stop

*example*

		./GoManga -mf "dokgo" 3-24

##### All together now

	./GoManga -mf "dokgo" 2 10-24 27 34 36-46




Downloads are kept in the users home directory in the format `Manga/{source}/{mangaName}/{chapter}`

### To be added/fixed:
 - [x] Add support for unordered chapter downloads, `1 2 45 22 56`
 - Fix failed downloads
 - Support updating all the mangas in the manga directory to the latest chapters
 - Support volume download
 - Threadpoolish implementation for downloads
 - Download into `.cbz` format or `.zip`
