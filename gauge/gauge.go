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

func NewPrometheusGauge(namespace, deploy, metricType string) (Gauge, error) {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mitose",
		Help: "Mitose autoscaller",
		ConstLabels: prometheus.Labels{
			"namespace":   namespace,
			"deploy":      deploy,
			"metric_type": metricType,
		},
	})
	if err := prometheus.Register(g); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}
	return &PrometheusGauge{pg: g}, nil
}

func Run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
