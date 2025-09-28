package datasource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/FIAP-SOAT-G20/hackathon-video-processor-job/internal/core/port"
)

// SNSMessageBroker implements storage operations using AWS S3
type SNSMessageBroker struct {
	client *sns.Client
	topic  string
}

// NewSnsMessageBroker creates a new S3 storage datasource
func NewSnsMessageBroker(client *sns.Client, topic string) port.MessageBroker {
	return &SNSMessageBroker{
		client: client,
		topic:  topic,
	}
}

func (ds *SNSMessageBroker) PublishMessage(ctx context.Context, message []byte) error {
	_, err := ds.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(ds.topic),
		Message:  aws.String(string(message)),
	})
	return err
}
