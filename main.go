package main

import (
	"context"
	"es_exporter/collector"
	"es_exporter/logger"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

func makePrometheusGathers(c prometheus.Collector) (prometheus.Gatherer, error) {
	registry := prometheus.NewRegistry()
	err := registry.Register(c)
	if err != nil {
		return nil, err
	}
	gathers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	return gathers, nil
}

func setHttpHandler(path string, gather prometheus.Gatherer) {
	http.Handle(path, promhttp.HandlerFor(gather, promhttp.HandlerOpts{}))
}

func activeCollector(path string, c prometheus.Collector) {
	gathers, err := makePrometheusGathers(c)
	if err != nil {
		logger.Logger.Warningf("path %s, collector failed to register %s", path, err)
	}
	setHttpHandler(path, gathers)
}

func main() {
	var (
		name              = "es_exporter"
		versionFlag       = flag.Bool("version", false, "show version")
		esURL             = flag.String("es.url", "http://localhost:9200", "Elasticsearch address")
		esTimeout         = flag.Duration("es.timeout", 10*time.Second, "Timeout for trying to get stats from elasticsearch")
		listenAddress     = flag.String("web.listen-address", ":8005", "Exporter listen on this address to push metric")
		healthMetricsPath = flag.String("web.health-metrics-path", "/health_metrics", "Path to expose health metrics")
		nodeMetricsPath   = flag.String("web.node-metrics-path", "/node_metrics", "Path to expose node metrics")
		indiceMetricsPath = flag.String("web.indice-metrics-path", "/indice_metrics", "Path to expose indices metrics")
	)
	flag.Parse()

	if *versionFlag {
		version.BuildUser = "gaoxun"
		version.BuildDate = "2017-12-20"
		version.Version = "0.0.1"
		fmt.Print(version.Print(name))
		os.Exit(0)
	}

	esAddress, err := url.Parse(*esURL)
	if err != nil {
		logger.Logger.Critical("es address is illegal", *esURL)
		os.Exit(-1)
	}

	esHttpClient := &http.Client{
		Timeout: *esTimeout,
	}

	// version metric, default
	// versionMetric := version.NewCollector(name)
	// prometheus.MustRegister(versionMetric)

	// Add collector here
	collectorMap := map[string]prometheus.Collector{
		*healthMetricsPath: collector.NewClusterHealth(esHttpClient, esAddress),
		*nodeMetricsPath:   collector.NewNodesStats(esHttpClient, esAddress),
		*indiceMetricsPath: collector.NewIndicesStats(esHttpClient, esAddress),
	}

	for path, c := range collectorMap {
		activeCollector(path, c)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			`<html>
		<head><title>Elasticsearch Exporter</title></head>
		<body>
			<h1>Elasticsearch Exporter</h1>
			<p><a href="` + *healthMetricsPath + `"> Health Metrics </a></p>
			<p><a href="` + *nodeMetricsPath + `"> Nodes Metrics </a></p>
			<p><a href="` + *indiceMetricsPath + `"> Indices Metrics </a></p>
		</body>
	</html>`))
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	h := &http.Server{Addr: *listenAddress, Handler: nil}

	logger.Logger.Info("Starting Http server...")
	go func() {
		if err := h.ListenAndServe(); err != nil {
			logger.Logger.Errorf("Http quit %s", err)
			stop <- os.Interrupt
		}
	}()

	<-stop
	logger.Logger.Info("Shutting down the server...")
	h.Shutdown(context.Background())
	logger.Logger.Info("Server gracefully stopped")
}
