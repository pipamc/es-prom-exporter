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

var (
	defaultIndexLabels      = []string{"index"}
	defaultIndexLabelValues = func(indexName string) []string {
		return []string{indexName}
	}
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
	subsystem := "indices"

	return &IndicesStats{
		client: client,
		url:    url,
		up:     common.UpMetric(namespace, subsystem, "cluster alive"),
		indexMetrics: []*indexMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "docs_primary"),
					"Count of documents with only primary shards",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Primaries.Docs.Count)
				},
				Labels: defaultIndexLabelValues,
			},

			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "store_size_bytes_primary"),
					"Current total size of stored index data in bytes with only primary shards on all nodes.",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Primaries.Store.SizeInBytes)
				},
				Labels: defaultIndexLabelValues,
			},

			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "store_size_bytes_total"),
					"Current total size of stored index data in bytes with all shards on all nodes.",
					defaultIndexLabels, nil,
				),
				Value: func(indexStats IndexStatsIndexResponse) float64 {
					return float64(indexStats.Total.Store.SizeInBytes)
				},
				Labels: defaultIndexLabelValues,
			},
		},
	}
}

func (c *IndicesStats) getIndicesMetric() (indexStatsResponse, error) {
	var isr indexStatsResponse

	u := *c.url
	u.Path = "/_all/_stats"
	res, err := c.client.Get(u.String())
	if err != nil {
		return isr, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return isr, common.HTTPStatusError(res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return isr, err
	}
	if err := json.Unmarshal(body, &isr); err != nil {
		return isr, err
	}
	return isr, nil
}

func (c *IndicesStats) Collect(ch chan<- prometheus.Metric) {
	defer func() {
		ch <- c.up
	}()

	isr, err := c.getIndicesMetric()
	if err != nil {
		c.up.Set(0)
		logger.Logger.Warningf("get indices info failed %s", err)
		return
	}
	c.up.Set(1)
	for indexName, indexStats := range isr.Indices {
		for _, metric := range c.indexMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(indexStats),
				metric.Labels(indexName)...,
			)
		}
	}
}

func (c *IndicesStats) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.indexMetrics {
		ch <- metric.Desc
	}
	ch <- c.up.Desc()
}
