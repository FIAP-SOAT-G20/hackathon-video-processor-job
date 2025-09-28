package port

import (
	"context"
)

// StorageDataSource defines the port for storage operations.
type MessageBroker interface {
	PublishMessage(ctx context.Context, message []byte) error
}
