package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/arizon-dread/plats/internal/model"
)

func CityFromZip(w http.ResponseWriter, r *http.Request) {

	zip := r.PathValue("zip")
	l := model.GetLocation(zip)

	if l.City != "" {
		writeOKResponse([]byte(l.City), w)
	} else {

	}

	conf := config.Load()
	//get all apiHosts
	apis := conf.Apis
	var channels = []chan string{}

	for i, api := range apis {
		channels = append(channels, make(chan string))
		go func() {
			err := getAddrFromApi(zip, api, channels[i])
			if err != nil {
				fmt.Printf("Got error when calling api, %v", err)
			}
		}()
	}
	cases := make([]reflect.SelectCase, len(channels))
	for i, ch := range channels {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	remaining := len(cases)
	found := false
	for remaining > 0 {
		chosen, val, ok := reflect.Select(cases)
		if !ok {
			//the channel is closed
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		}
		w.Write(val.Bytes())
		found = true
		cases[chosen].Chan.Close()
	}
	if !found {
		writeNotFound(w)
	}

	//Call address-api first, fmt.Sprintf(/api/v1/%v, r.PathValue("zip"))

}

func getAddrFromApi(zip string, api config.ApiHost, c chan<- string) error {
	if zip == "71897" {
		c <- "Dyltabruk"
		return nil
	}
	resp, err := http.Get(fmt.Sprintf("%v%v", api.Url, api.Path))
	if err != nil {
		return fmt.Errorf("error calling api url: %v, error was: %w", api.Url, err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body, %w", err)
	}
	city := model.City{}
	err = json.Unmarshal(b, &city)
	if err != nil {
		//how to handle failed parsing?
		return fmt.Errorf("error unmarshalling response body into city struct, %w", err)
	}
	c <- city.City
	return nil
}

func writeOKResponse(b []byte, w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write(b)
}

func writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
}
