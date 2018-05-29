package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/sohlich/majordomus/iflx"
)

func dataModule(c *iflx.InfluxClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/temp", withClient(c, writeTemp))
	return http.StripPrefix("/data", mux)
}

func withClient(cl *iflx.InfluxClient, handler IflxHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(cl, w, r)
	}
}

func writeTemp(client *iflx.InfluxClient, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sID := r.FormValue("sensorID")
	val := r.FormValue("val")
	log.Println(val)
	value, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "cannot read temperature", http.StatusInternalServerError)
	}
	log.Printf("Writing values %s,%.2f\n", sID, value)
	err = client.WriteTemp(sID, value)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "cannot write temperature", http.StatusInternalServerError)
	}
}
