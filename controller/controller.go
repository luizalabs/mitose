package controller

import (
	"context"
	"log"
	"time"

	"github.com/luizalabs/mitose/k8s"
)

const (
	numberOfMessagesInQueueAttrName       = "ApproximateNumberOfMessages"
	numberOfMessagesInFlightQueueAttrName = "ApproximateNumberOfMessagesNotVisible"
	msgsInQueueMetricName                 = "msgsInQueue"
	HPAScaleMethod                        = "HPA"
)

type Metrics map[string]string

type Colector interface {
	GetMetrics() (Metrics, error)
}

type Cruncher interface {
	CalcDesiredReplicas(Metrics) (int, error)
}

type Controller struct {
	colector    Colector
	cruncher    Cruncher
	namespace   string
	deployment  string
	scaleMethod string
	interval    time.Duration
}

func (c *Controller) Run(ctx context.Context) error {
	log.Printf("start controller for deployment %s (namespace %s)\n", c.deployment, c.namespace)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.interval):
			if err := c.Exec(); err != nil {
				return err
			}
		}
	}
}

func (c *Controller) Exec() error {
	m, err := c.colector.GetMetrics()
	if err != nil {
		return err
	}
	desiredReplicas, err := c.cruncher.CalcDesiredReplicas(m)
	if err != nil {
		return err
	}
	log.Printf(
		"Desired replicas %d for deployment %s (namespace %s)\n",
		desiredReplicas,
		c.deployment,
		c.namespace,
	)
	return c.Autoscale(desiredReplicas)
}

func (c *Controller) Autoscale(desiredReplicas int) error {
	if c.scaleMethod == HPAScaleMethod {
		return k8s.UpdateHPA(c.namespace, c.deployment, desiredReplicas, desiredReplicas)
	}
	return k8s.UpdateReplicasCount(c.namespace, c.deployment, desiredReplicas)
}

func NewController(colector Colector, cruncher Cruncher, namespace, deployment, scaleMethod, interval string) (*Controller, error) {
	convertedInterval, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}
	return &Controller{
		colector:    colector,
		cruncher:    cruncher,
		namespace:   namespace,
		deployment:  deployment,
		scaleMethod: scaleMethod,
		interval:    convertedInterval,
	}, nil
}
