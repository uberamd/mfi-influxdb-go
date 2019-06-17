package main

import (
	"encoding/json"
	"flag"
	"github.com/fatih/structs"
	"github.com/gorilla/mux"
	"github.com/influxdata/influxdb/client/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type BaseConfig struct {
	CheckInterval  int
	MfiAddr        string
	MfiUser        string
	MfiPass        string
	HealthHttpPort string
	InfluxDatabase string
	InfluxAddr     string
	InfluxUser     string
	InfluxPass     string
}

type MfiPowerSensors struct {
	Current     float64 `json:"current"`
	Voltage     float64 `json:"voltage"`
	PowerFactor float64 `json:"powerfactor"`
	ThisMonth   float64 `json:"thismonth"`
	Port        float64 `json:"port"`
	Output      float64 `json:"output"`
	Relay       float64 `json:"relay"`
	Lock        float64 `json:"lock"`
	PrevMonth   float64 `json:"prevmonth"`
	Power       float64 `json:"power"`
	Enabed      float64 `json:"enabled"`
	Label       string  `json:"label"`
}

type MfiPower struct {
	Sensors []MfiPowerSensors `json:"sensors"`
	Status  string            `json:"status"`
}

var (
	hostname, _ = os.Hostname()
)

func HealthHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

func pollMifiDevice(bc *BaseConfig) {
	for {
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr: bc.InfluxAddr,
			Username: bc.InfluxUser,
			Password: bc.InfluxPass,
		})

		if err != nil {
			log.Fatal(err)
		}


		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database: bc.InfluxDatabase,
			Precision: "s",
		})

		if err != nil {
			log.Fatal(err)
		}

		data := new(MfiPower)

		mfiEndpoint := bc.MfiAddr
		wc := http.Client{}
		form := url.Values{}
		form.Add("username", bc.MfiUser)
		form.Add("password", bc.MfiPass)
		req, err := http.NewRequest("POST", mfiEndpoint + "/login.cgi", strings.NewReader(form.Encode()))

		if err != nil {
			log.Printf("Error making POST: %v", err)
		}

		cookie := http.Cookie{
			Name: "AIROS_SESSIONID",
			Value: "01234567890123456789012345678901",
		}

		req.AddCookie(&cookie)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := wc.Do(req)
		if err != nil {
			log.Printf("Error with do post: %s", err)
		}

		resp.Body.Close()

		req, err = http.NewRequest("GET", mfiEndpoint + "/sensors", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Del("Content-Type")
		req.AddCookie(&cookie)
		resp, err = wc.Do(req)
		if err != nil {
			log.Printf("Error with get sensors: %s", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Fatal(err)
		}

		for _, d := range data.Sensors {
			tags := map[string]string{
				"mfi": bc.MfiAddr,
				"host": hostname,
				"label": d.Label,
			}

			pt, err := client.NewPoint("mfi_readings", tags, structs.Map(d), time.Now())
			if err != nil {
				log.Fatal(err)
			}

			bp.AddPoint(pt)
		}

		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		if err := c.Close(); err != nil {
			log.Fatal(err)
		}

		// cleanup http and influx clients
		resp.Body.Close()
		c.Close()

		log.Print("Points submitted to influxdb...")
		time.Sleep(20 * time.Second)
	}
}

func main() {
	bc := new(BaseConfig)
	flag.StringVar(&bc.MfiAddr,"mfi-addr", "http://127.0.0.1", "address of the Ubiquiti mFi device without trailing /")
	flag.StringVar(&bc.MfiUser,"mfi-user", "ubnt", "username of the Ubiquiti mFi device")
	flag.StringVar(&bc.MfiPass,"mfi-pass", "ubnt", "password of the Ubiquiti mFi device")
	flag.StringVar(&bc.HealthHttpPort,"http-port", "8085", "port for the http server to listen on for health checks")
	flag.StringVar(&bc.InfluxDatabase,"influxdb-database", "homelab_custom", "influxdb database to store datapoints")
	flag.StringVar(&bc.InfluxAddr,"influxdb-addr", "http://127.0.0.1:8086", "address of influxdb endpoint, ex: http://127.0.0.1:8086")
	flag.StringVar(&bc.InfluxUser,"influxdb-user", "admin", "username for influxdb access")
	flag.StringVar(&bc.InfluxPass,"influxdb-pass", "admin", "password for influxdb access")
	flag.IntVar(&bc.CheckInterval, "check-interval", 20, "frequency to poll the mFI for power data")

	flag.Parse()

	go pollMifiDevice(bc)

	r := mux.NewRouter()
	r.HandleFunc("/healthz", HealthHandler)
	log.Printf("Listening on :%s", bc.HealthHttpPort)

	log.Fatal(http.ListenAndServe(":" + bc.HealthHttpPort, r))
}
