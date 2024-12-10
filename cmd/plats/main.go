package main

import (
	"log"
	"net/http"

	"github.com/arizon-dread/plats/api/handler"
	"github.com/arizon-dread/plats/internal/config"
)

func main() {
	config := config.Load()
	//fail on no config
	if config.Cache.Url == "" {
		panic("could not read config")
	}
	//create stdlib http server
	mux := http.NewServeMux()

	//api endpoints
	mux.HandleFunc("GET /api/v1/zip/{zip}", handler.CityFromZip)

	//start api server
	log.Fatal(http.ListenAndServe(":8080", mux))
}
