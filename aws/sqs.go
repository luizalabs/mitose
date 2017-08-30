package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSClient struct {
	*AWS
}

func (s *SQSClient) GetQueueAttributes(queueURL string, attributes ...string) (map[string]string, error) {
	cli := sqs.New(session.New(), s.newConfig())

	var attrList []*string
	for _, attr := range attributes {
		a := attr
		attrList = append(attrList, &a)
	}

	out, err := cli.GetQueueAttributes(
		&sqs.GetQueueAttributesInput{QueueUrl: &queueURL, AttributeNames: attrList},
	)

	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range out.Attributes {
		result[k] = *v
	}
	return result, nil
}

func NewSQSClient(key, secret, region string) *SQSClient {
	return &SQSClient{
		AWS: &AWS{key: key, secret: secret, region: region},
	}
}
