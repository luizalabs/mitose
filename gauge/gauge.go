package gauge

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gauge interface {
	Set(float64) error
}

var (
	mu              *sync.Mutex
	registredGauges map[string]prometheus.Gauge
)

type PrometheusGauge struct {
	pg prometheus.Gauge
}

func (p *PrometheusGauge) Set(metric float64) error {
	p.pg.Set(metric)
	return nil
}

func init() {
	mu = new(sync.Mutex)
	registredGauges = make(map[string]prometheus.Gauge)
}

func NewPrometheusGauge(namespace, deploy, metricType string) Gauge {
	return &PrometheusGauge{pg: getOrCreateGauge(namespace, deploy, metricType)}
}

func getOrCreateGauge(namespace, deploy, metricType string) prometheus.Gauge {
	mu.Lock()
	defer mu.Unlock()

	gId := fmt.Sprintf("%s%s%s", namespace, deploy, metricType)
	if g, found := registredGauges[gId]; found {
		return g
	}
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mitose",
		Help: "Mitose autoscaller",
		ConstLabels: prometheus.Labels{
			"namespace":   namespace,
			"deploy":      deploy,
			"metric_type": metricType,
		},
	})
	prometheus.Register(g)
	registredGauges[gId] = g

	return g
}

func Run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
