package model

import (
	"github.com/arizon-dread/plats/internal/database"
)

type Location struct {
	Zip  string `json:"zipcode"`
	City string `json:"city"`
}

func (l *Location) Save(location Location) error {
	db := getImpl()
	return db.Store(location.Zip, location.City)
}

func GetLocation(key string) Location {
	var db = getImpl()
	val := db.Get(key)
	return Location{Zip: key, City: *val}
}
func getImpl() database.Db {
	return &database.Cache{}
}
