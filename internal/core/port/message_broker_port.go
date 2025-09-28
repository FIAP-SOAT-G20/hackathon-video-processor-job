package port

import (
	"context"
)

// MessageBroker defines the port for message broker operations, such as publishing messages.
type MessageBroker interface {
	PublishMessage(ctx context.Context, message []byte) error
}
