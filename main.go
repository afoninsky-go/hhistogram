package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afoninsky-go/hhistogram/processor"
	"github.com/afoninsky-go/logger"
	"github.com/gorilla/mux"
)

const httpTimeout = 15 * time.Second
const httpListenAddr = "127.0.0.1:8000"

func main() {
	log := logger.NewSTDLogger()
	router := mux.NewRouter()

	// convert incoming bulks of metrics into histograms
	router.Path("/bulk").Methods(http.MethodPost).HandlerFunc(MetricsHandler)

	// start http server
	srv := &http.Server{
		Handler:      router,
		Addr:         httpListenAddr,
		WriteTimeout: httpTimeout,
		ReadTimeout:  httpTimeout,
	}
	log.Infof("Running server on %s ...", httpListenAddr)
	log.Fatal(srv.ListenAndServe())
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := processor.NewHistogramConfig().WithName("test")
	bulk := processor.NewHistogramProcessor(*cfg)
	if err := bulk.AppendFromStream(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "OK")
}
