# GoManga
A little program to scrape/download manga off [mangareader] (http://www.mangareader.net/) and [mangafox] (http://www.mangafox.me/).
*Other sources might be added..eventually*

#Usage
Grab the executable in the bin folder and run it in the terminal..or compile the code if you Golang

The general usage format is
	
	./GoManga -source "manga name" chapter

valid sources are `-mf` for mangafox and `-mr` for mangareader

Download a single chapter.

    ./GoManga "manga name" chapter

If a source isn't provided, mangareader is used as the default

Download chapters over a certain range(advisable to keep the ranges short, <=50. mileage may vary)

    ./GoManga -source "manga name" start stop

Downloads are kept in the users home directory in the format `Manga/{source}/{mangaName}/{chapter}`
