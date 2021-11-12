package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Config struct {
	RefreshInterval int       `yaml:"refresh_interval"`
	Targets         []Targets `yaml:"targets"`
}

type Targets struct {
	As       int      `yaml:"as"`
	Prefixes []string `yaml:"prefixes"`
}

var addr = flag.String("listen-address", ":9959", "The address to listen on for HTTP requests.")
var metricsPath = flag.String("metrics-path", "/metrics", "Location where metrics should be exposed")
var configFile = flag.String("config-file", "config.yaml", "Configuration file location")
var config Config

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	flag.Parse()
	parseConfig()

	r := 3600
	if config.RefreshInterval > 0 {
		r = config.RefreshInterval
	}

	go func() {
		log.Info("Starting to collect metrics")
		for {
			collectMetrics()
			time.Sleep(time.Duration(r) * time.Second)
		}
	}()

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>RPKI Exporter</title></head>
             <body>
             <h1>RPKI Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func parseConfig() {
	cfgFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	err2 := yaml.Unmarshal(cfgFile, &config)
	if err2 != nil {
		log.Fatal(err2.Error())
	}

	validateConfig()

	log.Infof("Configuration file %v was parsed successfully \n", *configFile)
}

func validateConfig() {
	if len(config.Targets) == 0 {
		log.Fatal("No targets detected in the configuration file")
	}

	for _, c := range config.Targets {
		if !validateASN(c.As) {
			log.Fatal("AS Number in the configuration file is either invalid or not defined")
		}

		if len(c.Prefixes) == 0 {
			log.Fatalf("No prefixes defined for ASN: %v", c.As)
		}

		for _, prefix := range c.Prefixes {
			_, pNET, err := net.ParseCIDR(prefix)
			if err != nil || prefix != pNET.String() {
				log.Fatalf("Prefix is not valid: %v", prefix)
			}
		}
	}
}

func validateASN(asn int) bool {
	if (asn > 0) && (asn < 4200000000) {
		return true
	}
	return false
}
