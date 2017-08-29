package controller

import (
	"encoding/json"
	"math"
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
	QueueURLs  []string `json:"queue_urls"`
	MsgsPerPod int      `json:"msgs_per_pod"`
}

type SQSColector struct {
	queueURLs []string
	cli       *aws.SQSClient
}

type SQSCruncher struct {
	max        int
	min        int
	msgsPerPod int
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
	return Metrics{msgsInQueueMetricName: strconv.Itoa(msgsInQueue)}, nil
}

func (s *SQSColector) getNumberOfMsgsInQueue(queueURL string) (int, error) {
	attrs, err := s.cli.GetQueueAttributes(queueURL, numberOfMessageInQueueAttrName)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(attrs[numberOfMessageInQueueAttrName])
}

func (s *SQSCruncher) CalcDesiredReplicas(m Metrics) (int, error) {
	msgsInQueue, err := strconv.Atoi(m[msgsInQueueMetricName])
	if err != nil {
		return -1, err
	}
	desiredPods := float64(msgsInQueue) / float64(s.msgsPerPod)
	if desiredPods > float64(s.max) {
		return s.max, nil
	} else if desiredPods < float64(s.min) {
		return s.min, nil
	}
	return int(math.Ceil(desiredPods)), nil
}

func NewSQSColector(awsKey, awsSecret, awsRegion string, queueURLs ...string) Colector {
	cli := aws.NewSQSClient(awsKey, awsSecret, awsRegion)
	return &SQSColector{queueURLs: queueURLs, cli: cli}
}

func NewSQSCruncher(max, min, msgsPerPod int) Cruncher {
	return &SQSCruncher{max: max, min: min, msgsPerPod: msgsPerPod}
}

func NewSQSController(awsKey, awsSecret, awsRegion, confJSON string) (*Controller, error) {
	conf := new(SQSControlerConfig)
	if err := json.Unmarshal([]byte(confJSON), conf); err != nil {
		return nil, err
	}
	colector := NewSQSColector(awsKey, awsSecret, awsRegion, conf.QueueURLs...)
	cruncher := NewSQSCruncher(conf.Max, conf.Min, conf.MsgsPerPod)
	return NewController(colector, cruncher, conf.Namespace, conf.Deployment, conf.ScaleMethod), nil
}
