package main

import (
	"net/http"
	"time"

	"github.com/afoninsky-go/hhistogram/openapi"
	"github.com/afoninsky-go/hhistogram/processor"
	"github.com/afoninsky-go/logger"
	"github.com/gorilla/mux"
)

const httpTimeout = 15 * time.Second
const httpListenAddr = "127.0.0.1:8000"
const histogramName = "test"

func main() {
	log := logger.NewSTDLogger()
	processorConfig := processor.NewConfig().WithName(histogramName)
	router := mux.NewRouter()
	api := openapi.NewURLParser().WithLogger(log)

	// convert incoming bulks of metrics into histograms
	router.Path("/bulk").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create bulk processor to convert stream of http events into set of histograms
		bulk := processor.NewHistogramProcessor(processorConfig).WithInterceptor(api).WithLogger(log)
		// stream JSONP events
		if err := bulk.ReadFromStream(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// display result
		bulk.Process(w)
		// fmt.Fprintf(w, "OK")
	})

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
