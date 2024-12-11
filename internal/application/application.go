package application

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/arizon-dread/plats/internal/model"
)

func GetCity(zip string) []byte {

	//Creation of model checks the cache for a hit
	l := model.GetLocation(zip)

	if l.City != "" {
		return []byte(l.City)
	}

	conf := config.Load()
	//get all apiHosts
	apis := conf.Apis
	fallbacks := []config.ApiHost{}
	//remove fallbacks from the simultaneous calls
	apis = slices.DeleteFunc(apis, func(e config.ApiHost) bool {
		if e.Fallback {
			//add the fallbacks to their own slice
			fallbacks = append(fallbacks, e)
			return true
		}
		return false
	})
	city := getSimultaneously(zip, apis)
	//call fallbacks if we didn't find anything in the primary list and if we actually have items in the fallback slice
	if len(city) == 0 && len(fallbacks) > 0 {
		city = getSimultaneously(zip, fallbacks)
	}
	if len(city) > 0 {
		l.City = string(city)
		err := l.Save()
		if err != nil {
			fmt.Printf("couldn't cache %v, because %v\n", l, err)
		}
	}
	//return the city or an empty []byte{}
	return city

}

func getSimultaneously(zip string, apis []config.ApiHost) []byte {
	c := make(chan *string)
	timeout := make(chan bool)

	for _, api := range apis {
		//create a timeout goroutine
		go func() {
			time.Sleep(10 * time.Second)
			timeout <- true
		}()
		//create a go routine for each api
		go func() {
			err := getAddrFromApi(zip, api, c)
			if err != nil {
				fmt.Printf("Got error when calling api, %v\n", err)
			}
		}()
	}
	//Create a done var so we can count down until all api's have returned or timed out.
	done := len(apis)
	var value *string
	select {
	case <-c:
		value = <-c
		//only return and close c if we get an actual response. erroring in the goroutine will produce an empty string.
		if len(*value) > 0 {
			close(c)
			return []byte(*value)
		}
	case <-timeout:
		done--
		if done == 0 {
			close(timeout)
			close(c)
		}
	}
	return []byte{}
}

func getAddrFromApi(zip string, api config.ApiHost, c chan<- *string) error {
	if zip == "71897" {
		city := "Dyltabruk"
		c <- &city
		return nil
	}
	//Create an empty string that we send on the channel if an error occurs, so we can short-circuit the timeout.
	mtStr := ""
	resp, err := http.Get(fmt.Sprintf("%v%v", api.Url, api.Path))
	if err != nil {
		c <- &mtStr
		return fmt.Errorf("error calling api url: %v, error was: %w", api.Url, err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c <- &mtStr
		return fmt.Errorf("error reading response body, %w", err)
	}
	city := model.City{}
	err = json.Unmarshal(b, &city)
	if err != nil {
		c <- &mtStr
		return fmt.Errorf("error unmarshalling response body into city struct, %w", err)
	}
	c <- &city.City
	return nil
}
