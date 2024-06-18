package slo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/appclacks/server/pkg/slo/aggregates"
)

type Store interface {
	ListAggregatedRecords(ctx context.Context, threshold time.Time) ([]*aggregates.SLOSum, error)
	AddRecord(ctx context.Context, record aggregates.Record) error
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

func (s *Service) AddRecord(ctx context.Context, record aggregates.Record) error {
	return s.store.AddRecord(ctx, record)
}

func (s *Service) Metrics(ctx context.Context) error {
	sloSums, err := s.store.ListAggregatedRecords(ctx, time.Now().UTC().Add(-24*30*time.Hour))
	if err != nil {
		return err
	}
	for i := range sloSums {
		sloSum := sloSums[i]
		total := sloSum.Failure + sloSum.Success
		fmt.Println(total)

	}
	return nil
}
