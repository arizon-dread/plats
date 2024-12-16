package main

import (
	"log"
	"net/http"

	"github.com/arizon-dread/plats/api/handler"
	"github.com/arizon-dread/plats/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config := config.Load()
	//fail on no config
	if config.Cache.Url == "" {
		panic("could not read config")
	}
	//create stdlib http server
	mux := http.NewServeMux()

	//API ENDPOINTS

	//Utilities
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte("healthy")) })
	mux.Handle("GET /metrics", promhttp.Handler())
	//application endpoints
	mux.HandleFunc("GET /v1/zip/{zip}", handler.CityFromZip)

	//start api server
	log.Fatal(http.ListenAndServe(":8080", mux))
}
