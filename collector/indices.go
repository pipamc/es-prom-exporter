package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/url"
)

type indexMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(indexStats IndexStatsIndexResponse) float64
	Labels func(indexName string) []string
}

type IndicesStats struct {
	client       *http.Client
	url          *url.URL
	up           prometheus.Gauge
	indexMetrics []*indexMetric
}

func NewIndicesStats(client *http.Client, url *url.URL) *IndicesStats {
	return &IndicesStats{
		client: client,
		url:    url,
	}
}

func (c *IndicesStats) Collect(ch chan<- prometheus.Metric) {

}

func (c *IndicesStats) Describe(ch chan<- *prometheus.Desc) {

}
