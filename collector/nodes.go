package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/url"
)

type NodesStats struct {
	client *http.Client
	url    *url.URL
}

func NewNodesStats(client *http.Client, url *url.URL) *NodesStats {
	return &NodesStats{
		client: client,
		url:    url,
	}
}

func (c *NodesStats) Collect(ch chan<- prometheus.Metric) {

}

func (c *NodesStats) Describe(ch chan<- *prometheus.Desc) {

}
