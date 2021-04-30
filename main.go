package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/afoninsky-go/hhistogram/service"
	"github.com/afoninsky-go/logger"
	"github.com/gorilla/mux"
)

const httpTimeout = 15 * time.Second

func main() {
	log := logger.NewSTDLogger()

	httpListenAddr := fromEnv("HTTP_LISTEN", "127.0.0.1:8080")
	httpReceiver := fromEnv("HTTP_RECEIVER", "http://localhost:8081")
	hName := fromEnv("HISTOGRAM_NAME", "test_histogram")
	hDuration := fromEnv("HISTOGRAM_DURATION_MIN", "10")
	specFolder := fromEnv("SPEC_PATH", "./test/json")

	hDurationMin, err := strconv.Atoi(hDuration)
	log.FatalIfError(err)

	router := mux.NewRouter()
	histogram, err := service.NewHistogramService(service.Config{
		HistogramName:          hName,
		HistogramSliceDuration: time.Duration(hDurationMin) * time.Minute,
		OutputEndpoint:         httpReceiver,
		SpecFolder:             specFolder,
	})
	log.FatalIfError(err)

	router.Use(log.CreateMiddleware())
	router.Path("/bulk").Methods(http.MethodPost).HandlerFunc(histogram.BulkHandler)
	router.Path("/health").Methods(http.MethodGet, http.MethodHead).HandlerFunc(histogram.HealthHandler)
	log.CreateMiddleware()

	srv := &http.Server{
		Handler:      router,
		Addr:         httpListenAddr,
		WriteTimeout: httpTimeout,
		ReadTimeout:  httpTimeout,
	}
	log.Infof("Server on %s ...", httpListenAddr)
	log.Infof(`Generate histogram "%s" to %s ...`, hName, httpReceiver)
	log.Fatal(srv.ListenAndServe())
}

func fromEnv(envName, defValue string) string {
	envValue, ok := os.LookupEnv(envName)
	if ok {
		return envValue
	}
	return defValue
}
