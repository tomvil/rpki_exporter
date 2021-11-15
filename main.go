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
var metricsPath = flag.String("metrics-path", "/metrics", "Metrics location")
var configFile = flag.String("config-file", "config.yaml", "Configuration file location")
var config Config

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	flag.Parse()

	go func() {
		err := config.Parse()
		if err != nil {
			log.Fatal(err)
		}

		r := 3600
		if config.RefreshInterval > 0 {
			r = config.RefreshInterval
		}

		if config.Validate() {
			log.Info("Starting to collect metrics")
			for {
				collectMetrics()
				time.Sleep(time.Duration(r) * time.Second)
			}
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
	return nil
}

func (cfg Config) Validate() bool {
	if len(cfg.Targets) == 0 {
		log.Error("No targets detected in the configuration file")
		return false
	}

	for _, c := range cfg.Targets {
		if c.As <= 0 || c.As > 4200000000 {
			log.Fatal("AS Number in the configuration file is either invalid or not defined")
		}

		if len(c.Prefixes) == 0 {
			log.Errorf("No prefixes defined for ASN: %v", c.As)
			return false
		}

		for _, prefix := range c.Prefixes {
			_, pNET, err := net.ParseCIDR(prefix)
			if err != nil || prefix != pNET.String() {
				log.Fatalf("Prefix is not valid: %v", prefix)
			}
		}
	}
	return true
}
