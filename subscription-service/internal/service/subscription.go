package service

import (
	"context"
	"database/sql"
	"errors"
	"subscription-service/internal/db"
	"subscription-service/internal/model"

	"github.com/rs/zerolog"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionService interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id string) (*model.Subscription, error)
	List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id string) error
	Aggregate(ctx context.Context, from, to string, userID, serviceName *string) (int, error)
}

type subscriptionService struct {
	repo db.Repository
	log  *zerolog.Logger
}

func New(repo db.Repository, log *zerolog.Logger) SubscriptionService {
	return &subscriptionService{repo: repo, log: log}
}

// helper для указателей
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *subscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	s.log.Info().
		Str("user_id", sub.UserID).
		Str("service_name", sub.ServiceName).
		Int("price", sub.Price).
		Msg("Creating subscription")

	if err := s.repo.Create(ctx, sub); err != nil {
		s.log.Error().Err(err).Msg("repo create failed")
		return err
	}

	s.log.Debug().
		Str("id", sub.ID).
		Str("user_id", sub.UserID).
		Msg("Subscription created successfully")

	return nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	s.log.Info().Str("id", id).Msg("Fetching subscription by ID")

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn().Str("id", id).Msg("Subscription not found")
			return nil, ErrNotFound
		}
		s.log.Error().Err(err).Str("id", id).Msg("repo get failed")
		return nil, err
	}

	s.log.Debug().Str("id", sub.ID).Msg("Subscription fetched successfully")
	return sub, nil
}

func (s *subscriptionService) List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error) {
	s.log.Info().
		Str("user_id", userID).
		Str("service_name", serviceName).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Listing subscriptions")

	subs, err := s.repo.List(ctx, userID, serviceName, limit, offset)
	if err != nil {
		s.log.Error().Err(err).Msg("repo list failed")
		return nil, err
	}

	s.log.Debug().Int("count", len(subs)).Msg("Subscriptions listed successfully")
	return subs, nil
}

func (s *subscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	s.log.Info().
		Str("id", sub.ID).
		Str("user_id", sub.UserID).
		Str("service_name", sub.ServiceName).
		Int("price", sub.Price).
		Msg("Updating subscription")

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn().Str("id", sub.ID).Msg("Subscription not found")
			return ErrNotFound
		}
		s.log.Error().Err(err).Str("id", sub.ID).Msg("repo update failed")
		return err
	}

	s.log.Debug().Str("id", sub.ID).Msg("Subscription updated successfully")
	return nil
}

func (s *subscriptionService) Delete(ctx context.Context, id string) error {
	s.log.Info().Str("id", id).Msg("Deleting subscription")

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn().Str("id", id).Msg("Subscription not found")
			return ErrNotFound
		}
		s.log.Error().Err(err).Str("id", id).Msg("repo delete failed")
		return err
	}

	s.log.Debug().Str("id", id).Msg("Subscription deleted successfully")
	return nil
}

func (s *subscriptionService) Aggregate(ctx context.Context, from, to string, userID, serviceName *string) (int, error) {
	s.log.Info().
		Str("from", from).
		Str("to", to).
		Str("user_id", deref(userID)).
		Str("service_name", deref(serviceName)).
		Msg("Aggregating subscriptions")

	total, err := s.repo.AggregateTotal(ctx, from, to, userID, serviceName)
	if err != nil {
		s.log.Error().Err(err).Msg("repo aggregate failed")
		return 0, err
	}

	s.log.Debug().
		Str("from", from).
		Str("to", to).
		Int("total", total).
		Msg("Subscriptions aggregated successfully")

	return total, nil
}
