// Code generated by protoc-gen-go-messaging. DO NOT EDIT.

package github_pb

// Service: WebhookTopic
// Method: Push

func (msg *PushMessage) MessagingTopic() string {
	return "github-webhook"
}
func (msg *PushMessage) MessagingHeaders() map[string]string {
	headers := map[string]string{
		"grpc-service": "/o5.registry.github.v1.WebhookTopic/Push",
		"grpc-message": "o5.registry.github.v1.PushMessage",
		"api-version":  "d3da92b150f015554fc7ab24687173140b196813",
	}
	return headers
}
