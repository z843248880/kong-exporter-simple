package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"kong_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func createClientWithRetries(getClient func() (interface{}, error), retries uint, retryInterval time.Duration) (interface{}, error) {
	var err error
	var kongClient interface{}

	for i := 0; i <= int(retries); i++ {
		kongClient, err = getClient()
		if err == nil {
			return kongClient, nil
		}
		if i < int(retries) {
			log.Printf("Could not create Kong Client. Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}
	return nil, err
}

var (
	// Set during go build
	// version   string
	// gitCommit string

	// 命令行参数
	listenAddr       = flag.String("web.listen-port", "9001", "An port to listen on for web interface and telemetry.")
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	scrapeURI        = flag.String("kong.scrape-uri", "http://127.0.0.1:8001/status", "A URI for scraping Kong metrics.")
	metricsNamespace = flag.String("metric.namespace", "kong", "Prometheus metrics namespace, as the prefix of metrics name")
)

func main() {
	flag.Parse()
	httpClient := &http.Client{
		Timeout: time.Duration(7 * time.Second),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	// metrics := collector.NewMetrics(*metricsNamespace)
	registry := prometheus.NewRegistry()
	// registry.MustRegister(metrics)
	ossClient, err := createClientWithRetries(func() (interface{}, error) {
		return collector.NewKongClient(httpClient, *scrapeURI)
	}, 0, time.Duration(7 * time.Second))
	if err != nil {
		log.Fatalf("Could not create Kong Client: %v", err)
	}
	registry.MustRegister(collector.NewKongCollector(ossClient.(*collector.KongClient), "kong"))

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Kong Exporter</title></head>
			<body>
			<h1>Kong Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at :%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}
