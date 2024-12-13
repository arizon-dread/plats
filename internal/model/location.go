package model

import (
	"github.com/arizon-dread/plats/internal/config"
	"github.com/arizon-dread/plats/internal/database"
)

type Location struct {
	Zip  string `json:"zipcode"`
	City string `json:"city"`
}

func (l *Location) Save() error {
	db := getImpl()
	return db.Store(l.Zip, l.City)
}

func GetLocation(key string) Location {
	var db = getImpl()
	val, err := db.Get(key)
	if err != nil {
		return Location{Zip: key, City: ""}
	}
	return Location{Zip: key, City: val}
}
func getImpl() database.Db {
	conf := config.Load()
	if conf.Cache.Url != "" {
		return &database.Cache{}
	}
	return nil
}
