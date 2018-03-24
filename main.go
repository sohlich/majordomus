package main // import "github.com/sohlich/majordomus"

import (
	"net/http"

	"github.com/kelseyhightower/envconfig"

	"github.com/sohlich/majordomus/iflx"
)

var iflxCfg iflx.InflluxConfig

type IflxHandlerFunc func(client *iflx.InfluxClient, w http.ResponseWriter, r *http.Request)

func main() {
	envconfig.MustProcess("INFLUX", &iflxCfg)
	c := iflx.NewInfluxClient(iflxCfg)
	http.Handle("/data/", apiModule(c))
	http.ListenAndServe(":8080", nil)
}
