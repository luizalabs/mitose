package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	undeliveredMessagesMetric = "pubsub.googleapis.com/subscription/num_undelivered_messages"
	unackedMessagesMetric     = "pubsub.googleapis.com/subscription/num_outstanding_messages"
)

type PubSubClient struct {
	*GCPMetrics
}

func (m *GCPMetrics) GetNumOfUndeliveredMessages(subscriptionID string) (int, error) {
	c, err := m.newClient()
	if err != nil {
		return -1, err
	}
	ctx := context.Background()
	startTime := time.Now().UTC().Add(time.Minute * -1).Unix()
	endTime := time.Now().UTC().Unix()

	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/" + m.projectID,
		Filter: fmt.Sprintf(
			"(metric.type=\"%s\" or metric.type=\"%s\") AND resource.label.subscription_id=\"%s\"",
			undeliveredMessagesMetric, unackedMessagesMetric, subscriptionID,
		),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{Seconds: startTime},
			EndTime:   &timestamp.Timestamp{Seconds: endTime},
		},
	}
	iter := c.ListTimeSeries(ctx, req)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, fmt.Errorf("could not read time series value, %v ", err)
		}
		log.Printf("%+v\n", resp)
	}

	return 0, nil
}

func NewPubSubClient(googleApplicationCredentials, projectID, region string) *PubSubClient {
	return &PubSubClient{
		GCPMetrics: &GCPMetrics{
			googleApplicationCredentials: googleApplicationCredentials,
			region:                       region,
			projectID:                    projectID,
		},
	}
}
