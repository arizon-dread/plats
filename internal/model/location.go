package model

import (
	"github.com/arizon-dread/plats/internal/database"
)

type Location struct {
	Zip  string `json:"zipcode"`
	City string `json:"city"`
}

func (l *Location) Save(location Location) error {
	db := database.Db{}
	return db.Store(location.Zip, location.City)
}

func GetLocation(key string) Location {
	db := database.Db{}
	val := db.Get(key)
	return Location{Zip: key, City: *val}
}
