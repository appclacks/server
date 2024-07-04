package healthcheck

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/appclacks/server/internal/util"
	"github.com/appclacks/server/internal/validator"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	er "github.com/mcorbin/corbierror"
)

func InitDNSHealthcheck(healthcheck *aggregates.Healthcheck) {
	healthcheck.ID = util.NewUUID()
	healthcheck.CreatedAt = time.Now().UTC()
	healthcheck.Type = "dns"
	healthcheck.RandomID = rand.Intn(100000)
}

func InitCommandHealthcheck(healthcheck *aggregates.Healthcheck) {
	healthcheck.ID = util.NewUUID()
	healthcheck.CreatedAt = time.Now().UTC()
	healthcheck.Type = "command"
	healthcheck.RandomID = rand.Intn(100000)
}

func InitTCPHealthcheck(healthcheck *aggregates.Healthcheck) {
	healthcheck.ID = util.NewUUID()
	healthcheck.CreatedAt = time.Now().UTC()
	healthcheck.Type = "tcp"
	healthcheck.RandomID = rand.Intn(100000)
}

func InitHTTPHealthcheck(healthcheck *aggregates.Healthcheck) {
	healthcheck.ID = util.NewUUID()
	healthcheck.CreatedAt = time.Now().UTC()
	healthcheck.Type = "http"
	healthcheck.RandomID = rand.Intn(100000)
}
func InitTLSHealthcheck(healthcheck *aggregates.Healthcheck) {
	healthcheck.ID = util.NewUUID()
	healthcheck.CreatedAt = time.Now().UTC()
	healthcheck.Type = "tls"
	healthcheck.RandomID = rand.Intn(100000)
}

func (s *Service) CreateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error {
	s.logger.Info(fmt.Sprintf("creating healthcheck %s", healthcheck.Name))
	interval, err := time.ParseDuration(healthcheck.Interval)
	if err != nil {
		return er.New("Invalid healthcheck interval", er.BadRequest, true)
	}
	if interval < 30*time.Second {
		return er.New("The minimum healthcheck interval is 30 seconds", er.BadRequest, true)
	}
	timeout, err := time.ParseDuration(healthcheck.Timeout)
	if err != nil {
		return er.New("Invalid healthcheck timeout", er.BadRequest, true)
	}
	if interval < timeout {
		return er.New("The healthcheck interval should be greater than its timeout", er.BadRequest, true)
	}
	return s.store.CreateHealthcheck(ctx, healthcheck)
}

func (s *Service) UpdateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error {
	s.logger.Info(fmt.Sprintf("updating healthcheck %s", healthcheck.Name))
	err := validator.Validator.Struct(*healthcheck)
	if err != nil {
		return err
	}
	interval, err := time.ParseDuration(healthcheck.Interval)
	if err != nil {
		return er.New("Invalid healthcheck interval", er.BadRequest, true)
	}
	if interval < 2*time.Second {
		return er.New("The minimum healthcheck interval is 2 seconds", er.BadRequest, true)
	}
	timeout, err := time.ParseDuration(healthcheck.Timeout)
	if err != nil {
		return er.New("Invalid healthcheck timeout", er.BadRequest, true)
	}
	if interval < timeout {
		return er.New("The healthcheck interval should be greater than its timeout", er.BadRequest, true)
	}
	return s.store.UpdateHealthcheck(ctx, healthcheck)
}

func (s *Service) GetHealthcheck(ctx context.Context, id string) (*aggregates.Healthcheck, error) {
	return s.store.GetHealthcheck(ctx, id)
}

func (s *Service) GetHealthcheckByName(ctx context.Context, name string) (*aggregates.Healthcheck, error) {
	return s.store.GetHealthcheckByName(ctx, name)
}

func (s *Service) DeleteHealthcheck(ctx context.Context, id string) error {
	s.logger.Info(fmt.Sprintf("deleting healthcheck %s", id))
	return s.store.DeleteHealthcheck(ctx, id)
}

func (s *Service) ListHealthchecks(ctx context.Context, query aggregates.Query) ([]*aggregates.Healthcheck, error) {
	checks, err := s.store.ListHealthchecks(ctx, query.Enabled)
	if err != nil {
		return nil, err
	}
	if query.Regex == nil {
		return checks, nil
	}
	result := []*aggregates.Healthcheck{}
	for i := range checks {
		check := *checks[i]
		if query.Regex.MatchString(check.Name) {
			result = append(result, &check)
		}
	}
	return result, nil
}

func MatchLabels(healthcheck *aggregates.Healthcheck, labels map[string]string) bool {
	for labelKey, labelVal := range labels {
		val, ok := healthcheck.Labels[labelKey]
		if !ok {
			return false
		}
		if val != labelVal {
			return false
		}
	}
	return true
}

func (s *Service) CountHealthchecks(ctx context.Context) (int, error) {
	return s.store.CountHealthchecks(ctx)
}
