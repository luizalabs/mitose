package controller

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/gauge"
	"github.com/luizalabs/mitose/pubsub"
)

type PubSubControlerConfig struct {
	config.Config
	GoogleApplicationCredentials string   `json:"google_application_credentials"`
	Region                       string   `json:"region"`
	SubscriptionIDs              []string `json:"subscription_ids"`
	Project                      string   `json:"project"`
	MsgsPerPod                   int      `json:"msgs_per_pod"`
}

type PubSubColector struct {
	subscriptionIDs []string
	cli             *pubsub.PubSubClient
	gMetrics        gauge.Gauge
}

type PubSubCruncher struct {
	max        int
	min        int
	msgsPerPod int
	gMetrics   gauge.Gauge
}

func (s *PubSubColector) GetMetrics() (Metrics, error) {
	msgsInQueue := 0
	for _, queueURL := range s.subscriptionIDs {
		n, err := s.cli.GetNumOfUndeliveredMessages(queueURL)
		if err != nil {
			return nil, err
		}
		msgsInQueue += n
	}
	s.gMetrics.Set(float64(msgsInQueue))
	return Metrics{msgsInQueueMetricName: strconv.Itoa(msgsInQueue)}, nil
}

func (s *PubSubCruncher) CalcDesiredReplicas(m Metrics) (int, error) {
	desiredReplicas, err := s.calcReplicas(m)
	if err != nil {
		return -1, err
	}
	s.gMetrics.Set(float64(desiredReplicas))
	return desiredReplicas, nil
}

func (s *PubSubCruncher) calcReplicas(m Metrics) (int, error) {
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

func NewPubSubColector(g gauge.Gauge, googleApplicationCredentials, gcpProject, gcpRegion string, subscriptionIDs ...string) Colector {
	cli := pubsub.NewPubSubClient(googleApplicationCredentials, gcpProject, gcpRegion)
	return &PubSubColector{subscriptionIDs: subscriptionIDs, cli: cli, gMetrics: g}
}

func NewPubSubCruncher(g gauge.Gauge, max, min, msgsPerPod int) Cruncher {
	return &PubSubCruncher{max: max, min: min, msgsPerPod: msgsPerPod, gMetrics: g}
}

func NewPubSubController(confJSON string) (*Controller, error) {
	conf := new(PubSubControlerConfig)
	if err := json.Unmarshal([]byte(confJSON), conf); err != nil {
		return nil, err
	}

	gColector := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "PubSub")
	colector := NewPubSubColector(gColector, conf.GoogleApplicationCredentials, conf.Project, conf.Region, conf.SubscriptionIDs...)

	gCruncher := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "CRUNCHER")
	cruncher := NewPubSubCruncher(gCruncher, conf.Max, conf.Min, conf.MsgsPerPod)

	return NewController(
		colector,
		cruncher,
		conf.Namespace,
		conf.Deployment,
		conf.ScaleMethod,
		conf.Interval,
	)
}
