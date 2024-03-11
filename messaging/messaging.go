package messaging

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type SNSAPI interface {
	Publish(ctx context.Context, input *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type SNSPublisher struct {
	sns    SNSAPI
	prefix string
}

func NewSNSPublisher(sns SNSAPI, prefix string) *SNSPublisher {
	return &SNSPublisher{
		sns:    sns,
		prefix: prefix,
	}
}

type Message interface {
	MessagingTopic() string
	MessagingHeaders() map[string]string
	proto.Message
}

func (p *SNSPublisher) Publish(ctx context.Context, msg Message) error {
	protoBody, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	destination := msg.MessagingTopic()

	id := uuid.NewString()

	encoded := base64.StdEncoding.EncodeToString(protoBody)

	attributes := map[string]types.MessageAttributeValue{}

	headers := msg.MessagingHeaders()
	for key, val := range headers {
		if val == "" {
			continue
		}
		attributes[key] = types.MessageAttributeValue{
			StringValue: aws.String(val),
			DataType:    aws.String("String"),
		}
	}

	attributes["MessageId"] = types.MessageAttributeValue{
		StringValue: aws.String(id),
		DataType:    aws.String("String"),
	}

	dest := fmt.Sprintf("%s%s", p.prefix, destination)

	_, err = p.sns.Publish(ctx, &sns.PublishInput{
		Message:           aws.String(encoded),
		MessageAttributes: attributes,
		TopicArn:          aws.String(dest),
	})

	return err

}
