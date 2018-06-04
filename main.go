package main // import "github.com/sohlich/majordomus"

import (
	"flag"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/sohlich/majordomus/management"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/sohlich/majordomus/iflx"
)

const CONNSTRING = "postgres://pgtest:pgtest@localhost:5432/majordomus?sslmode=disable"
const DRIVER = "postgres"

var jwtSecret string
var domain string

var iflxCfg iflx.InflluxConfig

type IflxHandlerFunc func(client *iflx.InfluxClient, w http.ResponseWriter, r *http.Request)

func main() {

	flag.StringVar(&jwtSecret, "jwt", "devel", "")
	flag.StringVar(&domain, "domain", "localhost", "")
	flag.Parse()

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
		UserProperty: "token",
	})

	db, err := sqlx.Connect(DRIVER, CONNSTRING)
	if err != nil {
		panic(err)
	}

	m := management.NewGeneralStore(management.MgmtModuleCfg{db, jwtSecret, domain})
	// envconfig.MustProcess("INFLUX", &iflxCfg)
	// c := iflx.NewInfluxClient(iflxCfg)
	// http.Handle("/data/", dataModule(c))
	http.Handle("/auth/",
		http.StripPrefix("/auth", m.ApiAuthModule()))
	http.Handle("/group/",
		jwtMiddleware.Handler(
			http.StripPrefix("/group", m.ApiGroupModule())))
	http.Handle("/device/",
		jwtMiddleware.Handler(
			http.StripPrefix("/device", m.ApiDeviceModule())))
	http.ListenAndServe(":8080", nil)
}
