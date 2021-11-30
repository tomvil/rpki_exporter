package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	Data Data `json:"validated_route"`
}

type Data struct {
	Route    Route
	Validity Validity
}

type Route struct {
	OriginAsn string `json:"origin_asn"`
	Prefix    string
}

type Validity struct {
	State string
}

var rpkiStatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "rpki_status",
		Help: "RPKI Status of the prefix (0 - invalid, 1 - valid, 2 - not found)",
	}, []string{"prefix", "asn"})

var rpkiQueriesFailedTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rpki_queries_failed_total",
		Help: "Number of failed queries",
	})

var rpkiQueriesSuccessTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rpki_queries_success_total",
		Help: "Number of successful queries",
	})

var status = map[string]float64{
	"invalid":   0,
	"valid":     1,
	"not-found": 2,
}

func init() {
	prometheus.MustRegister(rpkiStatus)
	prometheus.MustRegister(rpkiQueriesSuccessTotal)
	prometheus.MustRegister(rpkiQueriesFailedTotal)
}

func collectMetrics() {
	for _, target := range config.Targets {
		for _, prefix := range target.Prefixes {
			go setPrefixRPKIStatus(prefix, target.As)
		}
	}
}

func setPrefixRPKIStatus(prefix string, as int) {
	var responseObject Response

	url := fmt.Sprintf("https://rpki-validator.ripe.net/validity?asn=%v&prefix=%v", as, prefix)

	responseData, err := requestGET(url)
	if err != nil {
		rpkiQueriesFailedTotal.Inc()
		log.Error(err)

		return
	}

	err2 := json.Unmarshal(responseData, &responseObject)
	if err2 != nil {
		log.Fatalf("Failed to unmarshal response: %v", err2)
	}

	rpkiStatus.WithLabelValues(
		responseObject.Data.Route.Prefix,
		responseObject.Data.Route.OriginAsn).Set(status[responseObject.Data.Validity.State])

	rpkiQueriesSuccessTotal.Inc()
}

func requestGET(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v status returned: %v", url, response.StatusCode)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}
