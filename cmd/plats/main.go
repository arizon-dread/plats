package main

import (
	"net/http"

	"github.com/arizon-dread/plats/api/handler"
	"github.com/arizon-dread/plats/internal/config"
)

func main() {
	config := config.Config{}
	config.Load()
	//fail on no config
	if config.Cache.Url == "" {
		panic("could not read config")
	}
	//create stdlib http server
	mux := http.NewServeMux()

	//api endpoints
	mux.HandleFunc("GET /api/v1/zip/{zip}", handler.CityFromZip)

	//start api server
	http.ListenAndServe(":8080", mux)
}
