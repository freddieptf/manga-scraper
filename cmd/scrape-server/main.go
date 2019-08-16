package main

import (
	"fmt"
	"github.com/freddieptf/manga-scraper/pkg/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultServerPort = "8001"
)

func main() {
	fmt.Printf("server running at localhost:%s\n", defaultServerPort)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	compressor := middleware.NewCompressor(5, "text/plain")
	router.Use(compressor.Handler())
	router.Route("/manga", func(r chi.Router) {
		sources := api.GetMangaSources()
		for _, source := range sources {
			r.Mount(fmt.Sprintf("/%s", strings.ToLower(source.Name())), api.GetSourceRouter(source))
		}
	})
	srv := &http.Server{
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		Addr:         fmt.Sprintf(":%s", defaultServerPort),
	}
	log.Fatal(srv.ListenAndServe())
}
