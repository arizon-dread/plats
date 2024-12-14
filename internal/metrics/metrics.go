package metrics

import (
	"fmt"
	"sync"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metrix Metrix
var once sync.Once

type ApiMetrix struct {
	Name    string
	ApiHits prometheus.Counter
}
type Metrix struct {
	CacheHits prometheus.Counter
	ApiMetrix []ApiMetrix
}

func GetMetrics() *Metrix {
	once.Do(func() {
		metrix = Metrix{}
		conf := config.Load()
		for _, api := range conf.Apis {
			metrix.ApiMetrix = append(metrix.ApiMetrix, ApiMetrix{Name: api.Name, ApiHits: promauto.NewCounter(prometheus.CounterOpts{
				Name: api.Name + "_hits_total",
				Help: fmt.Sprintf("The number of hits for %v api", api.Name),
			})})
		}
		metrix.CacheHits = promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "The number of cache hits for this instance",
		})

	})

	return &metrix
}
