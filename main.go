package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ip is used for storing the return data
// from the ipify api endpoint
type ip struct {
	Ip string
}

// background updates the IP in a background process
// every 10 seconds
func background(m *metrics) {
	t := time.NewTicker(10 * time.Second)
	currentIP := ""

	for _ = range t.C {
		newIp, err := getIP()
		fmt.Println(newIp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if newIp != currentIP {
			//deregister old vector
			m.ip.Delete(prometheus.Labels{"address": currentIP})
			//register new vector
			m.ip.With(prometheus.Labels{"address": newIp}).Set(float64(1))
			currentIP = newIp
		}
	}
}

// getIP queries a public api that returns your IP address
// and returns it as a string
func getIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var r ip
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}
	return r.Ip, nil
}

type metrics struct {
	ip *prometheus.GaugeVec
}

// registerMetrics sets up the metrics that will be returned
// by the /metrics endpoint
func registerMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		ip: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "node_public_ip",
				Help: "Current public facing IP address of a node",
			},
			[]string{"address"},
		),
	}
	reg.MustRegister(m.ip)
	return m
}

var (
	port string
	path string
)

func init() {
	port = os.Getenv("EXPORTER_PORT")
	path = os.Getenv("EXPORTER_PATH")
	if port == "" {
		port = ":8080"
	}
	if path == "" {
		path = "/metrics"
	}
	fmt.Printf("port: %s\npath: %s\n", port, path)
}

func main() {
	reg := prometheus.NewRegistry()
	m := registerMetrics(reg)
	go background(m)
	http.Handle(path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(port, nil))
}
