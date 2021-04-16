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

func main() {
	log := logger.NewSTDLogger()
	router := mux.NewRouter()
	api := openapi.NewURLParser()

	// convert incoming bulks of metrics into histograms
	router.Path("/bulk").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := processor.NewHistogramConfig().WithName("test")
		bulk := processor.NewHistogramProcessor(*cfg).WithInterceptor(api)
		if err := bulk.ReadFromStream(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
