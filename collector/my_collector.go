package collector

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// 指标结构体
type Metrics struct {
	kongClient *KongClient
	metrics    map[string]*prometheus.Desc
	upMetric   prometheus.Gauge
	mutex      sync.Mutex
}

/**
 * 工厂方法：NewMetrics
 * 功能：初始化指标信息，即Metrics结构体
 */
func NewKongCollector(kongClient *KongClient, namespace string) *Metrics {
	return &Metrics{
		kongClient: kongClient,
		metrics: map[string]*prometheus.Desc{
			"connections_active":   newGlobalMetric(namespace, "connections_active", "Active client connections"),
			"connections_accepted": newGlobalMetric(namespace, "connections_accepted", "Accepted client connections"),
			"connections_handled":  newGlobalMetric(namespace, "connections_handled", "Handled client connections"),
			"connections_reading":  newGlobalMetric(namespace, "connections_reading", "Connections where NGINX is reading the request header"),
			"connections_writing":  newGlobalMetric(namespace, "connections_writing", "Connections where NGINX is writing the response back to the client"),
			"connections_waiting":  newGlobalMetric(namespace, "connections_waiting", "Idle client connections"),
			"http_requests_total":  newGlobalMetric(namespace, "http_requests_total", "Total http requests"),
		},
		upMetric: newUpMetric(namespace),
	}
}

/**
 * 接口：Describe
 * 功能：传递结构体中的指标描述符到channel
 */
func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()

	for _, m := range c.metrics {
		ch <- m
	}
}

/**
 * 接口：Collect
 * 功能：抓取最新的数据，传递给channel
 */
func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects
	defer c.mutex.Unlock()

	stats, err := c.kongClient.GetStubStats()
	if err != nil || stats.Database.Reachable != true {
		c.upMetric.Set(kongDown)
		ch <- c.upMetric
		log.Printf("Error getting stats: %v", err)
		return
	}

	c.upMetric.Set(kongUp)
	ch <- c.upMetric

	ch <- prometheus.MustNewConstMetric(c.metrics["connections_active"],
		prometheus.GaugeValue, stats.Server.ConnectionsActive)
	ch <- prometheus.MustNewConstMetric(c.metrics["connections_accepted"],
		prometheus.CounterValue, stats.Server.ConnectionsAccepted)
	ch <- prometheus.MustNewConstMetric(c.metrics["connections_handled"],
		prometheus.CounterValue, stats.Server.ConnectionsHandled)
	ch <- prometheus.MustNewConstMetric(c.metrics["connections_reading"],
		prometheus.GaugeValue, stats.Server.ConnectionsReading)
	ch <- prometheus.MustNewConstMetric(c.metrics["connections_writing"],
		prometheus.GaugeValue, stats.Server.ConnectionsWriting)
	ch <- prometheus.MustNewConstMetric(c.metrics["connections_waiting"],
		prometheus.GaugeValue, stats.Server.ConnectionsWaiting)
	ch <- prometheus.MustNewConstMetric(c.metrics["http_requests_total"],
		prometheus.CounterValue, stats.Server.TotalRequests)

}
