package heartbeat

import (
	"context"
	"log/slog"

	"github.com/appclacks/server/pkg/heartbeat/aggregates"
)

type Store interface {
	CreateHeartbeat(ctx context.Context, heartbeat *aggregates.Heartbeat) error
	GetHeartbeat(ctx context.Context, id string) (*aggregates.Heartbeat, error)
	GetHeartbeatByName(ctx context.Context, name string) (*aggregates.Heartbeat, error)
	DeleteHeartbeat(ctx context.Context, id string) error
	ListHeartbeats(ctx context.Context) ([]*aggregates.Heartbeat, error)
	UpdateHeartbeat(ctx context.Context, heartbeat *aggregates.Heartbeat) error
	RefreshHeartbeat(ctx context.Context, id string) error
	CountHeartbeats(ctx context.Context) (int, error)
}

type Service struct {
	logger *slog.Logger
	store  Store
}

func New(logger *slog.Logger, store Store) *Service {
	return &Service{
		logger: logger,
		store:  store,
	}
}