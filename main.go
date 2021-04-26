package main

import (
	"net/http"
	"time"

	"github.com/afoninsky-go/hhistogram/service"
	"github.com/afoninsky-go/logger"
	"github.com/gorilla/mux"
)

const httpTimeout = 15 * time.Second
const httpListenAddr = "127.0.0.1:8080"

func main() {
	log := logger.NewSTDLogger()
	router := mux.NewRouter()
	histogram, err := service.NewHistogramService(service.Config{
		HistogramName:  "test",
		OutputEndpoint: "http://localhost:8081",
		SpecFolder:     "./test/json",
	})
	log.FatalIfError(err)

	router.Use(log.CreateMiddleware())
	router.Path("/bulk").Methods(http.MethodPost).HandlerFunc(histogram.BulkHandler)
	router.Path("/health").Methods(http.MethodGet).HandlerFunc(histogram.HealthHandler)
	log.CreateMiddleware()

	srv := &http.Server{
		Handler:      router,
		Addr:         httpListenAddr,
		WriteTimeout: httpTimeout,
		ReadTimeout:  httpTimeout,
	}
	log.Infof("Running server on %s ...", httpListenAddr)
	log.Fatal(srv.ListenAndServe())
}
