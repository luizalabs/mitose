package controller

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/gauge"
	"github.com/luizalabs/mitose/rabbitmq"
)

type RabbitMQControlerConfig struct {
	config.Config
	Credentials string   `json:"credentials"`
	QueueURLs   []string `json:"queue_urls"`
	MsgsPerPod  int      `json:"msgs_per_pod"`
}

type RabbitMQColector struct {
	QueueURLs   []string
	Credentials string
	Request     *rabbitmq.RabbitMQClient
	gMetrics    gauge.Gauge
}

type RabbitMQCruncher struct {
	max        int
	min        int
	msgsPerPod int
	gMetrics   gauge.Gauge
}

func (s *RabbitMQColector) GetMetrics() (Metrics, error) {
	msgsInQueue := 0
	for _, queueURL := range s.QueueURLs {
		n, err := s.Request.GetNumOfMessages(queueURL, s.Credentials)
		if err != nil {
			return nil, err
		}
		msgsInQueue += n
	}
	s.gMetrics.Set(float64(msgsInQueue))
	return Metrics{msgsInQueueMetricName: strconv.Itoa(msgsInQueue)}, nil
}

func (s *RabbitMQCruncher) CalcDesiredReplicas(m Metrics) (int, error) {
	desiredReplicas, err := s.calcReplicas(m)
	if err != nil {
		return -1, err
	}
	s.gMetrics.Set(float64(desiredReplicas))
	return desiredReplicas, nil
}

func (s *RabbitMQCruncher) calcReplicas(m Metrics) (int, error) {
	msgsInQueue, err := strconv.Atoi(m[msgsInQueueMetricName])
	if err != nil {
		return -1, err
	}
	desiredReplicas := float64(msgsInQueue) / float64(s.msgsPerPod)
	if desiredReplicas > float64(s.max) {
		return s.max, nil
	} else if desiredReplicas < float64(s.min) {
		return s.min, nil
	}
	desiredReplicas = math.Ceil(desiredReplicas)
	return int(desiredReplicas), nil
}

func NewRabbitMQColector(g gauge.Gauge, credentials string, queueURLs ...string) Colector {
	return &RabbitMQColector{QueueURLs: queueURLs, Credentials: credentials, gMetrics: g}
}

func NewRabbitMQCruncher(g gauge.Gauge, max, min, msgsPerPod int) Cruncher {
	return &RabbitMQCruncher{max: max, min: min, msgsPerPod: msgsPerPod, gMetrics: g}
}

func NewRabbitMQController(confJSON string) (*Controller, error) {
	conf := new(RabbitMQControlerConfig)
	if err := json.Unmarshal([]byte(confJSON), conf); err != nil {
		return nil, err
	}

	gColector := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "RabbitMQ")
	colector := NewRabbitMQColector(gColector, conf.Credentials, conf.QueueURLs...)

	gCruncher := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "CRUNCHER")
	cruncher := NewRabbitMQCruncher(gCruncher, conf.Max, conf.Min, conf.MsgsPerPod)

	return NewController(
		colector,
		cruncher,
		conf.Namespace,
		conf.Deployment,
		conf.ScaleMethod,
		conf.Interval,
	)
}
