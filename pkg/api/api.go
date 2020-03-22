package api

import (
	"encoding/json"
	"github.com/freddieptf/manga-scraper/pkg/mangareader"
	"github.com/freddieptf/manga-scraper/pkg/scraper"
	"github.com/go-chi/chi"
	"net/http"
)

type MangaSource interface {
	Name() string
	Search(mangaName string) ([]scraper.Manga, error)
	ListMangaDirectory() ([]scraper.Manga, error)
	ListMangaChapters(mangaID string) ([]scraper.Chapter, error)
	GetChapter(mangaID, chapterID string) (scraper.Chapter, error)
}

func GetMangaSources() []MangaSource {
	sources := []MangaSource{}
	sources = append(sources, &mangareader.ReaderManga{})
	return sources
}

func GetSourceRouter(source MangaSource) http.Handler {
	return chi.NewRouter().Route("/", func(r chi.Router) {
		r.Get("/directory", listMangaDirectory(source))
		r.Route("/{mangaID}", func(r chi.Router) {
			r.Get("/chapters", listMangaChapters(source))
			r.Get("/{chapterID}", getChapter(source))
		})
	})
}

func listMangaDirectory(source MangaSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mangaList, err := source.ListMangaDirectory()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		body, err := json.Marshal(mangaList)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(body))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func listMangaChapters(source MangaSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chapterList, err := source.ListMangaChapters(chi.URLParam(r, "mangaID"))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		body, err := json.Marshal(chapterList)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(body))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func getChapter(source MangaSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chapter, err := source.GetChapter(chi.URLParam(r, "mangaID"), chi.URLParam(r, "chapterID"))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		body, err := json.Marshal(chapter)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", http.DetectContentType(body))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}
