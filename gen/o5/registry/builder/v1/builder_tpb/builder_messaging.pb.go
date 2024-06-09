// Code generated by protoc-gen-go-messaging. DO NOT EDIT.

package builder_tpb

// Service: BuilderRequestTopic
// Method: BuildProto

func (msg *BuildProtoMessage) MessagingTopic() string {
	return "registry-build_request"
}
func (msg *BuildProtoMessage) MessagingHeaders() map[string]string {
	headers := map[string]string{
		"grpc-service": "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildProto",
		"grpc-message": "o5.registry.builder.v1.topic.BuildProtoMessage",
	}
	if msg.Request != nil {
		headers["o5-request-reply-to"] = msg.Request.ReplyTo
	}
	return headers
}

// Method: BuildAPI

func (msg *BuildAPIMessage) MessagingTopic() string {
	return "registry-build_request"
}
func (msg *BuildAPIMessage) MessagingHeaders() map[string]string {
	headers := map[string]string{
		"grpc-service": "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildAPI",
		"grpc-message": "o5.registry.builder.v1.topic.BuildAPIMessage",
	}
	if msg.Request != nil {
		headers["o5-request-reply-to"] = msg.Request.ReplyTo
	}
	return headers
}

// Service: BuilderReplyTopic
// Method: BuildStatus

func (msg *BuildStatusMessage) MessagingTopic() string {
	return "github-status_reply"
}
func (msg *BuildStatusMessage) MessagingHeaders() map[string]string {
	headers := map[string]string{
		"grpc-service": "/o5.registry.builder.v1.topic.BuilderReplyTopic/BuildStatus",
		"grpc-message": "o5.registry.builder.v1.topic.BuildStatusMessage",
	}
	if msg.Request != nil {
		headers["o5-reply-reply-to"] = msg.Request.ReplyTo
	}
	return headers
}
