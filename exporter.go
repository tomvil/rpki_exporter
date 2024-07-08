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
	VRPs  VRPs
}

type VRPs struct {
	Matched         []VRP `json:"matched"`
	UnmatchedLength []VRP `json:"unmatched_length"`
}

type VRP struct {
	MaxLength string `json:"max_length"`
}

var rpkiStatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "rpki_status",
		Help: "RPKI Status of the prefix (0 - invalid, 1 - valid, 2 - not found)",
	}, []string{"prefix", "asn", "max_length", "unmatched_length"})

var rpkiQueriesFailedTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rpki_queries_failed_total",
		Help: "Number of failed queries",
	})

var rpkiQueriesSuccessfulTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rpki_queries_successful_total",
		Help: "Number of successful queries",
	})

var status = map[string]float64{
	"invalid":   0,
	"valid":     1,
	"not-found": 2,
}

func init() {
	prometheus.MustRegister(rpkiStatus)
	prometheus.MustRegister(rpkiQueriesSuccessfulTotal)
	prometheus.MustRegister(rpkiQueriesFailedTotal)
}

func collectMetrics() {
	for _, target := range config.Targets {
		for _, prefix := range target.Prefixes {
			go setPrefixRPKIStatus(prefix, *target.As)
		}
	}
}

func setPrefixRPKIStatus(prefix string, as uint) {
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

	maxLength := "NOT FOUND"
	if len(responseObject.Data.Validity.VRPs.Matched) > 0 {
		maxLength = responseObject.Data.Validity.VRPs.Matched[0].MaxLength
	}

	unmatchedLength := "0"
	if len(responseObject.Data.Validity.VRPs.UnmatchedLength) > 0 {
		unmatchedLength = "1"
	}

	rpkiStatus.WithLabelValues(
		responseObject.Data.Route.Prefix,
		responseObject.Data.Route.OriginAsn,
		maxLength,
		unmatchedLength).Set(status[responseObject.Data.Validity.State])

	rpkiQueriesSuccessfulTotal.Inc()
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
