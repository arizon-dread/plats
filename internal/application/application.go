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
	"sync"
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
			log.Printf("couldn't cache %v, because %v\n", l, err)
		}
	}
	//return the city or an empty []byte{}
	return city

}

func getSimultaneously(zip string, apis []config.ApiHost) []byte {
	result := make(chan *string)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	for _, api := range apis {
		wg.Add(1)
		//create a go routine for each api
		go func() {
			err := getAddrFromApi(zip, api, ctx, result, &wg)
			if err != nil {
				log.Printf("Got error when calling api, %v\n", err)
				return
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

func getAddrFromApi(zip string, api config.ApiHost, ctx context.Context, c chan<- *string, wg *sync.WaitGroup) error {
	defer wg.Done()
	if zip == "71897" {
		city := "Dyltabruk"
		c <- &city
		return nil
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	errs := errors.New("")
	req, err := http.NewRequestWithContext(reqCtx, "GET", fmt.Sprintf("%v%v", api.Url, api.Path), nil)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("got error when creating http request, %w", err))
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err == nil {
		b, err := io.ReadAll(resp.Body)
		if err == nil {
			city := model.City{}
			err = json.Unmarshal(b, &city)
			if err == nil {
				// if the context is done, another routine has finished the call and this is trailing after, then we just return. otherwise, write to the channel
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(5 * time.Second):
					emptyStr := ""
					if &city.City != &emptyStr {
						c <- &city.City
						ctx.Done()
						return nil
					}
					return fmt.Errorf("timeout was reached")
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
