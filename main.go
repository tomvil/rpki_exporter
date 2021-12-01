package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RefreshInterval int `yaml:"refresh_interval"`
	Targets         []Targets
}

type Targets struct {
	As       int
	Prefixes []string
}

var addr = flag.String("web.listen-address", ":9959", "The address to listen on for HTTP requests.")
var metricsPath = flag.String("web.metrics-path", "/metrics", "Metrics location")
var configFile = flag.String("config.file", "config.yaml", "Configuration file location")
var config Config

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	flag.Parse()

	go func() {
		err := config.Parse()
		if err != nil {
			log.Fatal(err)
		}

		err2 := config.Validate()
		if err2 != nil {
			log.Error(err2)

			return
		}

		log.Info("Starting to collect metrics")

		for {
			collectMetrics()
			time.Sleep(time.Duration(config.RefreshInterval) * time.Second)
		}

	}()

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<html>
             <head><title>RPKI Exporter</title></head>
             <body>
             <h1>RPKI Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`)); err != nil {
			log.Fatal(err)
		}
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func (cfg *Config) Parse() error {
	cfgFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}

	err2 := yaml.Unmarshal(cfgFile, &cfg)
	if err2 != nil {
		return err2
	}

	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = 600
	}

	return nil
}

func (cfg Config) Validate() error {
	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no targets detected in the configuration file")
	}

	for _, target := range cfg.Targets {
		if target.As <= 0 || target.As > 4200000000 {
			return fmt.Errorf(
				"AS Number in the configuration file is either invalid or not defined",
			)
		}

		if len(target.Prefixes) == 0 {
			return fmt.Errorf("no prefixes defined for ASN: %v", target.As)
		}

		for _, prefix := range target.Prefixes {
			_, pNET, err := net.ParseCIDR(prefix)
			if err != nil || prefix != pNET.String() {
				return fmt.Errorf("prefix is not valid: %v", prefix)
			}
		}
	}

	return nil
}
