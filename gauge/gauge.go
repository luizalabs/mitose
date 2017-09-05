package gauge

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Gauge interface {
	Set(float64) error
}

type PrometheusGauge struct {
	pg prometheus.Gauge
}

func (p *PrometheusGauge) Set(metric float64) error {
	p.pg.Set(metric)
	return nil
}

func NewPrometheusGauge(namespace, deploy, metricType string) Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mitose",
		Help: "Mitose autoscaller",
		ConstLabels: prometheus.Labels{
			"namespace":   namespace,
			"deploy":      deploy,
			"metric_type": metricType,
		},
	})
	prometheus.MustRegister(g)
	return &PrometheusGauge{pg: g}
}

func Run() error {
	port := os.Getenv("PORT")
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
