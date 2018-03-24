package iflx

import (
	"fmt"
	"log"

	client "github.com/influxdata/influxdb/client/v2"
)

type InfluxClient struct {
	c   client.Client
	cfg InflluxConfig
}

type InflluxConfig struct {
	HOST     string `default:"localhost"`
	PORT     string `default:"8086"`
	Database string `default:"telegraf"`
	Username string `default:"telegraf"`
	Password string `default:"telegraf"`
}

func (cfg *InflluxConfig) String() string {
	return fmt.Sprintf(`
		URL:%s
		Database:%s
		Username:%s
		`, cfg.HOST+":"+cfg.PORT, cfg.Database, cfg.Username)
}

func NewInfluxClient(cfg InflluxConfig) *InfluxClient {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://" + cfg.HOST + ":" + cfg.PORT,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	log.Println("Initializing Influx client")
	log.Println(cfg.String())
	return &InfluxClient{c, cfg}
}

func (ic *InfluxClient) WriteTemp(sensorId string, temp float64) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  ic.cfg.Database,
		Precision: "s",
	})

	tags := map[string]string{"sensor_id": sensorId}
	fields := map[string]interface{}{"temp": temp}
	p, err := client.NewPoint("temperature", tags, fields)
	if err != nil {
		return err
	}

	bp.AddPoint(p)
	return ic.c.Write(bp)
}
