[![pipeline status](https://gitlab.com/freddieptf/manga-scraper/badges/master/pipeline.svg)](https://gitlab.com/freddieptf/manga-scraper/commits/master)


## Usage

The general usage format is

	./scraper --{source} --manga "manga name" chapters

*example*

	./scraper --mf "dokgo" 2


##### Download chapters over a certain range

	./scraper --{source} --manga "manga name" start-stop

*example*

	./scraper --mf --manga "dokgo" 2 3-24 56


##### Downloading volumes (only on mangafox)
This will download volume 1 to 5 of dokgo

	./scraper --mf --vlm "dokgo" 1-5


#### Need some quick Help?

	./scraper --help


Downloads are kept in the users home directory in the format `~/Manga/{source}/{mangaName}/{chapter}`
