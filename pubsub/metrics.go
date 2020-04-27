package pubsub

import (
	"context"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/option"
)

type GCPMetrics struct {
	googleApplicationCredentials string
	projectID                    string
	region                       string
}

func (m *GCPMetrics) newClient() (*monitoring.MetricClient, error) {
	ctx := context.Background()
	return monitoring.NewMetricClient(ctx, option.WithCredentialsFile(m.googleApplicationCredentials))
}
