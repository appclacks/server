package heartbeat

import (
	"context"
	"fmt"
	"time"

	"github.com/appclacks/server/internal/util"
	"github.com/appclacks/server/internal/validator"
	"github.com/appclacks/server/pkg/heartbeat/aggregates"
	er "github.com/mcorbin/corbierror"
)

func InitHeartbeat(heartbeat *aggregates.Heartbeat) {
	heartbeat.ID = util.NewUUID()
	heartbeat.CreatedAt = time.Now().UTC()
}

func (s *Service) CreateHeartbeat(ctx context.Context, heartbeat *aggregates.Heartbeat) error {
	s.logger.Info(fmt.Sprintf("creating heartbeat %s", heartbeat.Name))

	if heartbeat.TTL != nil {
		_, err := time.ParseDuration(*heartbeat.TTL)
		if err != nil {
			return er.New("invalid heartbeat TTL", er.BadRequest, true)
		}
	}

	err := validator.Validator.Struct(*heartbeat)
	if err != nil {
		return err
	}

	return s.store.CreateHeartbeat(ctx, heartbeat)
}

func (s *Service) UpdateHeartbeat(ctx context.Context, heartbeat *aggregates.Heartbeat) error {
	s.logger.Info(fmt.Sprintf("updating heartbeat %s", heartbeat.Name))

	err := validator.Validator.Struct(*heartbeat)
	if err != nil {
		return err
	}

	if heartbeat.TTL != nil {
		_, err := time.ParseDuration(*heartbeat.TTL)
		if err != nil {
			return er.New("invalid heartbeat TTL", er.BadRequest, true)
		}
	}

	return s.store.UpdateHeartbeat(ctx, heartbeat)
}

func (s *Service) GetHeartbeat(ctx context.Context, id string) (*aggregates.Heartbeat, error) {
	return s.store.GetHeartbeat(ctx, id)
}

func (s *Service) GetHeartbeatByName(ctx context.Context, name string) (*aggregates.Heartbeat, error) {
	return s.store.GetHeartbeatByName(ctx, name)
}

func (s *Service) DeleteHeartbeat(ctx context.Context, id string) error {
	s.logger.Info(fmt.Sprintf("deleting heartbeat %s", id))
	return s.store.DeleteHeartbeat(ctx, id)
}

func (s *Service) ListHeartbeats(ctx context.Context) ([]*aggregates.Heartbeat, error) {
	return s.store.ListHeartbeats(ctx)
}

func (s *Service) RefreshHeartbeat(ctx context.Context, id string) error {
	s.logger.Info(fmt.Sprintf("refreshing heartbeat %s", id))
	return s.store.RefreshHeartbeat(ctx, id)
}

func (s *Service) CountHeartbeats(ctx context.Context) (int, error) {
	return s.store.CountHeartbeats(ctx)
}

func MatchLabels(heartbeat *aggregates.Heartbeat, labels map[string]string) bool {
	for labelKey, labelVal := range labels {
		val, ok := heartbeat.Labels[labelKey]
		if !ok {
			return false
		}
		if val != labelVal {
			return false
		}
	}
	return true
}
