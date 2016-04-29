# GoManga
A little program to scrape/download manga off [mangareader] (http://www.mangareader.net/) and [mangafox] (http://www.mangafox.me/).
*Other sources might be added..eventually*

##Usage
Grab the executable in the bin folder and run it in the terminal..or compile the source if you Golang

The general usage format is

	./GoManga -source "manga name" chapter

valid sources are `-mf` for mangafox and `-mr` for mangareader

Download a single chapter.

    ./GoManga "manga name" chapter

*If a source isn't provided, mangareader is used as the default*

Download chapters over a certain range(advisable to keep the ranges short, <=50. for now.)

    ./GoManga -source "manga name" start stop

Downloads are kept in the users home directory in the format `Manga/{source}/{mangaName}/{chapter}`

###To be added/fixed:
 - Add support for unordered chapter downloads, `1 2 45 22 56`
 - Fix failed downloads
 - Support updating all the mangas in the manga directory to the latest chapters
 - Support volume download
 - Threadpoolish implementation for downloads
 - Download into `.cbz` format or `.zip`
