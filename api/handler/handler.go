package handler

import (
	"fmt"
	"net/http"

	"github.com/arizon-dread/plats/internal/model"
)

func CityFromZip(w http.ResponseWriter, r *http.Request) {

	zip := r.PathValue("zip")
	l := model.GetLocation(zip)

	if l.City != "" {
		writeOKResponse([]byte(l.City), w)
	} else if city, err := getFromAddrApi(zip); err == nil {
		writeOKResponse([]byte(city), w)
	} else if city, err = getFromExternalApi(zip); err == nil {
		writeOKResponse([]byte(city), w)
	} else {
		writeNotFound(w)
	}

	//Call address-api first, fmt.Sprintf(/api/v1/%v, r.PathValue("zip"))

}

func getFromAddrApi(zip string) (string, error) {
	if zip == "12340" {
		return "Farsta", nil
	}

	return "", fmt.Errorf("not found")
}

func getFromExternalApi(zip string) (string, error) {
	if zip == "12340" {
		return "Farsta", nil
	}

	return "", fmt.Errorf("not found")
}

func writeOKResponse(b []byte, w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write(b)
}

func writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
}
