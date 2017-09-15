package controller

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/luizalabs/mitose/aws"
	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/gauge"
)

const (
	numberOfMessagesInQueueAttrName       = "ApproximateNumberOfMessages"
	numberOfMessagesInFlightQueueAttrName = "ApproximateNumberOfMessagesNotVisible"
	msgsInQueueMetricName                 = "msgsInQueue"
)

type SQSControlerConfig struct {
	config.Config
	Key        string   `json:"key"`
	Secret     string   `json:"secret"`
	Region     string   `json:"region"`
	QueueURLs  []string `json:"queue_urls"`
	MsgsPerPod int      `json:"msgs_per_pod"`
}

type SQSColector struct {
	queueURLs []string
	cli       *aws.SQSClient
	gMetrics  gauge.Gauge
}

type SQSCruncher struct {
	max        int
	min        int
	msgsPerPod int
	gMetrics   gauge.Gauge
}

func (s *SQSColector) GetMetrics() (Metrics, error) {
	msgsInQueue := 0
	for _, queueURL := range s.queueURLs {
		n, err := s.getNumberOfMsgsInQueue(queueURL)
		if err != nil {
			return nil, err
		}
		msgsInQueue += n
	}
	s.gMetrics.Set(float64(msgsInQueue))
	return Metrics{msgsInQueueMetricName: strconv.Itoa(msgsInQueue)}, nil
}

func (s *SQSColector) getNumberOfMsgsInQueue(queueURL string) (int, error) {
	attrs, err := s.cli.GetQueueAttributes(
		queueURL,
		numberOfMessagesInQueueAttrName,
		numberOfMessagesInFlightQueueAttrName,
	)
	if err != nil {
		return -1, err
	}
	visible, err := strconv.Atoi(attrs[numberOfMessagesInQueueAttrName])
	if err != nil {
		return -1, err
	}
	inFlight, err := strconv.Atoi(attrs[numberOfMessagesInFlightQueueAttrName])
	if err != nil {
		return -1, err
	}
	return visible + inFlight, nil
}

func (s *SQSCruncher) CalcDesiredReplicas(m Metrics) (int, error) {
	desiredReplicas, err := s.calcReplicas(m)
	if err != nil {
		return -1, err
	}
	s.gMetrics.Set(float64(desiredReplicas))
	return desiredReplicas, nil
}

func (s *SQSCruncher) calcReplicas(m Metrics) (int, error) {
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

func NewSQSColector(g gauge.Gauge, awsKey, awsSecret, awsRegion string, queueURLs ...string) Colector {
	cli := aws.NewSQSClient(awsKey, awsSecret, awsRegion)
	return &SQSColector{queueURLs: queueURLs, cli: cli, gMetrics: g}
}

func NewSQSCruncher(g gauge.Gauge, max, min, msgsPerPod int) Cruncher {
	return &SQSCruncher{max: max, min: min, msgsPerPod: msgsPerPod, gMetrics: g}
}

func NewSQSController(confJSON string) (*Controller, error) {
	conf := new(SQSControlerConfig)
	if err := json.Unmarshal([]byte(confJSON), conf); err != nil {
		return nil, err
	}

	gColector := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "SQS")
	colector := NewSQSColector(gColector, conf.Key, conf.Secret, conf.Region, conf.QueueURLs...)

	gCruncher := gauge.NewPrometheusGauge(conf.Namespace, conf.Deployment, "CRUNCHER")
	cruncher := NewSQSCruncher(gCruncher, conf.Max, conf.Min, conf.MsgsPerPod)

	return NewController(
		colector,
		cruncher,
		conf.Namespace,
		conf.Deployment,
		conf.ScaleMethod,
		conf.Interval,
	)
}
