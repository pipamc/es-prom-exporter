package common

import "github.com/prometheus/client_golang/prometheus"

func UpMetric(namespace string, subsystem string, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName(namespace, subsystem, "up"),
		Help: help,
	})
}
