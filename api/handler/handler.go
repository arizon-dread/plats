package handler

import (
	"net/http"

	"github.com/arizon-dread/plats/internal/application"
)

func CityFromZip(w http.ResponseWriter, r *http.Request) {

	zip := r.PathValue("zip")
	city := application.GetCity(zip)
	if len(city) > 0 {
		writeOKResponse(city, w)
		return
	}
	writeNotFound(w)
	//Call address-api first, fmt.Sprintf(/api/v1/%v, r.PathValue("zip"))

}

func writeOKResponse(b []byte, w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write(b)
}

func writeNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
}
