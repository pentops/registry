// Code generated by protoc-gen-go-o5-messaging. DO NOT EDIT.
// versions:
// - protoc-gen-go-o5-messaging 0.0.0
// source: j5/registry/github/v1/topic/webhook.proto

package github_tpb

import (
	context "context"
	o5msg "github.com/pentops/o5-messaging/o5msg"
)

// Service: WebhookTopic
type WebhookTopicTxSender[C any] struct {
	sender o5msg.TxSender[C]
}

func NewWebhookTopicTxSender[C any](sender o5msg.TxSender[C]) *WebhookTopicTxSender[C] {
	sender.Register(o5msg.TopicDescriptor{
		Service: "j5.registry.github.v1.topic.WebhookTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "Push",
				Message: (*PushMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &WebhookTopicTxSender[C]{sender: sender}
}

type WebhookTopicCollector[C any] struct {
	collector o5msg.Collector[C]
}

func NewWebhookTopicCollector[C any](collector o5msg.Collector[C]) *WebhookTopicCollector[C] {
	collector.Register(o5msg.TopicDescriptor{
		Service: "j5.registry.github.v1.topic.WebhookTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "Push",
				Message: (*PushMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &WebhookTopicCollector[C]{collector: collector}
}

type WebhookTopicPublisher struct {
	publisher o5msg.Publisher
}

func NewWebhookTopicPublisher(publisher o5msg.Publisher) *WebhookTopicPublisher {
	publisher.Register(o5msg.TopicDescriptor{
		Service: "j5.registry.github.v1.topic.WebhookTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "Push",
				Message: (*PushMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &WebhookTopicPublisher{publisher: publisher}
}

// Method: Push

func (msg *PushMessage) O5MessageHeader() o5msg.Header {
	header := o5msg.Header{
		GrpcService:      "j5.registry.github.v1.topic.WebhookTopic",
		GrpcMethod:       "Push",
		Headers:          map[string]string{},
		DestinationTopic: "github-webhook",
	}
	return header
}

func (send WebhookTopicTxSender[C]) Push(ctx context.Context, sendContext C, msg *PushMessage) error {
	return send.sender.Send(ctx, sendContext, msg)
}

func (collect WebhookTopicCollector[C]) Push(sendContext C, msg *PushMessage) {
	collect.collector.Collect(sendContext, msg)
}

func (publish WebhookTopicPublisher) Push(ctx context.Context, msg *PushMessage) {
	publish.publisher.Publish(ctx, msg)
}