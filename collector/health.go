package collector

import (
	"encoding/json"
	"es_exporter/common"
	"es_exporter/logger"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch"
)

var (
	colors                     = [...]string{"green", "yellow", "red"}
	defaultClusterHealthLabels = []string{"cluster"}
)

type clusterHealthMetric struct {
	Type     prometheus.ValueType
	Describe *prometheus.Desc
	Value    func(clusterHealth clusterHealthResponse) float64
	Labels   func(clusterName string) []string
}

type clusterStatusMetric struct {
	Type     prometheus.ValueType
	Describe *prometheus.Desc
	Value    func(clusterStatus clusterHealthResponse, status string) float64
	Labels   func(clusterName string, color string) []string
}

type ClusterHealth struct {
	client *http.Client
	url    *url.URL
	up     prometheus.Gauge

	metrics      []*clusterHealthMetric
	statusMetric *clusterStatusMetric
}

func NewClusterHealth(client *http.Client, url *url.URL) *ClusterHealth {
	subsystem := "cluster_health"

	return &ClusterHealth{
		client: client,
		url:    url,

		up: common.UpMetric(namespace, subsystem, "cluster alive"),

		metrics: []*clusterHealthMetric{
			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "active_primary_shards"),
					"The number of active primary shards.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.ActivePrimaryShards)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "active_shards"),
					"The total number of all active shards.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.ActiveShards)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "delayed_unassigned_shards"),
					"Shards delayed to reduce reallocation overhead.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.DelayedUnassignedShards)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "initializing_shards"),
					"Count of shards that are being freshly created.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.InitializingShards)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "number_of_data_nodes"),
					"Number of data nodes in the cluster.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.NumberOfDataNodes)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "number_of_in_flight_fetch"),
					"The number of ongoing shard info requests.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.NumberOfInFlightFetch)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "number_of_nodes"),
					"Number of nodes in the cluster.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.NumberOfNodes)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "number_of_pending_tasks"),
					"Cluster level changes which have not yet been executed.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.NumberOfPendingTasks)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "relocating_shards"),
					"The number of shards that are currently moving from one node to another node.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.RelocatingShards)
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "timed_out"),
					"Number of cluster health checks timed out.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					if clusterHealth.TimedOut {
						return 1
					}
					return 0
				},
			},

			{
				Type: prometheus.GaugeValue,
				Describe: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "unassigned_shards"),
					"The number of shards that exist in the cluster state, but cannot be found in the cluster itself.",
					defaultClusterHealthLabels, nil,
				),
				Value: func(clusterHealth clusterHealthResponse) float64 {
					return float64(clusterHealth.UnassignedShards)
				},
			},
		},

		statusMetric: &clusterStatusMetric{
			Type: prometheus.GaugeValue,
			Describe: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "status"),
				"Whether all shards are allocated",
				[]string{"cluster", "color"},
				nil),
			Value: func(clusterStatus clusterHealthResponse, status string) float64 {
				if clusterStatus.Status == status {
					return 1
				}
				return 0
			},
		},
	}
}

func (c *ClusterHealth) getMetrics() (clusterHealthResponse, error) {
	var chr clusterHealthResponse

	u := *c.url
	u.Path = "/_cluster/health"
	res, err := c.client.Get(u.String())
	if err != nil {
		return chr, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return chr, common.HTTPStatusError(res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return chr, err
	}
	if err := json.Unmarshal(body, &chr); err != nil {
		return chr, err
	}
	return chr, nil
}

func (c *ClusterHealth) Collect(ch chan<- prometheus.Metric) {
	defer func() {
		ch <- c.up
	}()

	chr, err := c.getMetrics()
	if err != nil {
		c.up.Set(0)
		logger.Logger.Warningf("Can not get metric from elasticsearch, err %s", err)
		return
	}
	c.up.Set(1)
	for _, color := range colors {
		ch <- prometheus.MustNewConstMetric(
			c.statusMetric.Describe,
			c.statusMetric.Type,
			c.statusMetric.Value(chr, color),
			chr.ClusterName, color,
		)
	}
}

func (c *ClusterHealth) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.Describe
	}
	ch <- c.statusMetric.Describe
	ch <- c.up.Desc()
}
