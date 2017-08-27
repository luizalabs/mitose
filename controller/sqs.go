package controller

import (
	"encoding/json"
	"strconv"

	"github.com/luizalabs/mitose/aws"
	"github.com/luizalabs/mitose/config"
)

const (
	numberOfMessageInQueueAttrName = "ApproximateNumberOfMessages"
	msgsInQueueMetricName          = "msgsInQueue"
)

type SQSControlerConfig struct {
	config.Config
	QueueURL   string `json:"queue_url"`
	MsgsPerPod int    `json:"msgs_per_pod"`
}

type SQSColector struct {
	queueURL string
	cli      *aws.SQSClient
}

type SQSCruncher struct {
	max        int
	min        int
	msgsPerPod int
}

func (s *SQSColector) GetMetrics() (Metrics, error) {
	attrs, err := s.cli.GetQueueAttributes(s.queueURL, numberOfMessageInQueueAttrName)
	if err != nil {
		return nil, err
	}
	return Metrics{
		msgsInQueueMetricName: attrs[numberOfMessageInQueueAttrName],
	}, nil
}

func (s *SQSCruncher) CalcDesiredReplicas(m Metrics) (int, error) {
	msgsInQueue, err := strconv.Atoi(m[msgsInQueueMetricName])
	if err != nil {
		return -1, err
	}
	desiredPods := msgsInQueue / s.msgsPerPod
	if desiredPods > s.max {
		return s.max, nil
	} else if desiredPods < s.min {
		return s.min, nil
	}
	return desiredPods, nil
}

func NewSQSColector(awsKey, awsSecret, awsRegion, queueURL string) Colector {
	cli := aws.NewSQSClient(awsKey, awsSecret, awsRegion)
	return &SQSColector{queueURL: queueURL, cli: cli}
}

func NewSQSCruncher(max, min, msgsPerPod int) Cruncher {
	return &SQSCruncher{max: max, min: min, msgsPerPod: msgsPerPod}
}

func NewSQSController(awsKey, awsSecret, awsRegion, confJSON string) (*Controller, error) {
	conf := new(SQSControlerConfig)
	if err := json.Unmarshal([]byte(confJSON), conf); err != nil {
		return nil, err
	}
	colector := NewSQSColector(awsKey, awsSecret, awsRegion, conf.QueueURL)
	cruncher := NewSQSCruncher(conf.Max, conf.Min, conf.MsgsPerPod)
	return NewController(colector, cruncher, conf.Namespace, conf.Deployment), nil
}
