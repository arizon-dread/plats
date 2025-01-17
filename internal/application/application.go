package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/arizon-dread/plats/internal/metrics"
	"github.com/arizon-dread/plats/internal/model"
	"github.com/tidwall/gjson"
)

func GetCity(zip string) []byte {
	mtx := metrics.GetMetrics()

	//Creation of model checks the cache for a hit
	l := model.GetLocation(zip)

	if l.City != "" {
		mtx.CacheHits.Inc()
		return []byte(l.City)
	}

	conf := config.Load()
	//get all apiHosts into a cloned list so we can filter out the fallbacks w/o removing them from the config singleton
	apis := slices.Clone(conf.Apis)
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
	city := getSimultaneously(zip, &apis)
	//call fallbacks if we didn't find anything in the primary list and if we actually have items in the fallback slice
	if len(city) == 0 && len(fallbacks) > 0 {
		city = getSimultaneously(zip, &fallbacks)
	}
	if len(city) > 0 {
		l.City = string(city)
		err := l.Save()
		if err != nil {
			log.Printf("couldn't cache %v, because %v\n", l, err)
		}
	}
	//return the city or an empty []byte{}
	return city

}

func getSimultaneously(zip string, apis *[]config.ApiHost) []byte {
	result := make(chan *string)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	for _, api := range *apis {
		wg.Add(1)
		//create a go routine for each api
		go func() {
			defer wg.Done()
			err := getAddrFromApi(zip, &api, ctx, result, &wg)
			if err != nil && len(err.Error()) > 0 {
				log.Printf("Got error when calling api, %v\n", err)
			}
		}()
	}
	//we can wait in a different routine for the fan-out to finish
	go func() {
		wg.Wait()
		close(result)
	}()
	//read from the channel
	select {
	case res, ok := <-result:
		cancel()
		//if the channel produced a value, return it as a []byte
		if ok && res != nil {
			return []byte(*res)
		}
		//if we have a closed channel and an empty result, return an empty []byte{} and produce a 404 in the api
		return []byte{}
	case <-ctx.Done():
		//the context finished but there was no response
		cancel()
		return []byte{}
	}
}

func getAddrFromApi(zip string, api *config.ApiHost, ctx context.Context, c chan<- *string, wg *sync.WaitGroup) error {
	if zip == "71897" {
		city := "Dyltabruk"
		c <- &city
		return nil
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	path := api.Path
	path = strings.Replace(path, "${zip}", zip, -1)
	path = strings.Replace(path, "${apikey}", api.ApiKey, -1)

	errs := errors.New("")
	req, err := http.NewRequestWithContext(reqCtx, "GET", fmt.Sprintf("%v%v", api.Url, path), nil)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("got error when creating http request, %w", err))
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err == nil {
		b, err := io.ReadAll(resp.Body)
		if err == nil {
			//Log response headers if specified, to track rate quota.
			for _, v := range api.LogHeaders {
				if resp.Header.Get(v) != "" {
					log.Printf("%v.%v: %v\n", api.Name, v, resp.Header.Get(v))
				}
			}

			city := model.City{}
			//if we have a gjson-readable path to where the "city" is in the response, use that to parse the response and set the city value.
			if len(api.ResponseCityKey) > 0 {
				c := gjson.Get(string(b), api.ResponseCityKey).String()
				if len(c) > 0 {
					city.City = c
				} else {
					err = fmt.Errorf("unable to parse json response with gjson and api.ResponseCityKey: %v", api.ResponseCityKey)
				}
			} else {
				err = json.Unmarshal(b, &city)
			}
			if err == nil {
				// if the context is done, another routine has finished the call and this is trailing after, then we just return. otherwise, write to the channel
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(5 * time.Second):
					wg.Add(1)
					writeToChan(city.City, api, c, ctx, wg)
					return fmt.Errorf("timeout was reached")
				default:
					wg.Add(1)
					writeToChan(city.City, api, c, ctx, wg)
				}
			} else {
				errs = errors.Join(errs, fmt.Errorf("error unmarshalling response body into city struct, %w", err))
			}
		} else {
			errs = errors.Join(errs, fmt.Errorf("error reading response body, %w", err))
		}
	} else {
		errs = errors.Join(errs, fmt.Errorf("error calling api url: %v, error was: %w", api.Url, err))
	}
	return errs
}
func writeToChan(city string, api *config.ApiHost, c chan<- *string, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if city != "" {
		c <- &city
		go addMetric(api)
		ctx.Done()
	}
}

func addMetric(api *config.ApiHost) {

	mtx := metrics.GetMetrics()
	for i, apiHit := range mtx.ApiMetrix {

		if strings.Contains(apiHit.Name, api.Name) {
			//Increase metric for the api that got to serve the response.
			mtx.ApiMetrix[i].ApiHits.Inc()
		}
	}
}
