package service

import (
	"REST_service/internal/models"
	"REST_service/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SubscriptionService interface {
	Create(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) ([]models.Subscription, int64, error)
	GetTotalCost(ctx context.Context, req *models.TotalCostRequest) (*models.TotalCostResponse, error)
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *logrus.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *logrus.Logger) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

var (
    ErrInvalidDateRange = errors.New("end_date cannot be before start_date")
    ErrSubscriptionNotFound = errors.New("subscription not found")
)

func (s *subscriptionService) Create(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"price":        req.Price,
	}).Info("creating new subscription")

	if req.EndDate != nil && req.EndDate.Time.Before(req.StartDate.Time) {
		return nil, ErrInvalidDateRange
	}

	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate.Time,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.EndDate != nil {
		sub.EndDate = &req.EndDate.Time
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.WithError(err).Error("failed to create subscription")
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.WithField("subscription_id", sub.ID).Info("subscription created successfully")
	return sub, nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	s.logger.WithField("subscription_id", id).Debug("fetching subscription by ID")

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("failed to get subscription")
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	return sub, nil
}

func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	s.logger.WithField("subscription_id", id).Info("updating subscription")

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).Error("subscription not found")
		return nil, ErrSubscriptionNotFound
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		sub.StartDate = req.StartDate.Time
	}
	if req.EndDate != nil {
		sub.EndDate = &req.EndDate.Time
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return nil, ErrInvalidDateRange
	}

	sub.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, sub); err != nil {
		s.logger.WithError(err).Error("failed to update subscription")
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.WithField("subscription_id", id).Info("subscription updated successfully")
	return sub, nil
}

func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.WithField("subscription_id", id).Info("deleting subscription")

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.WithError(err).Error("subscription not found")
			return ErrSubscriptionNotFound
		}
		s.logger.WithError(err).Error("failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.logger.WithField("subscription_id", id).Info("subscription deleted successfully")
	return nil
}

func (s *subscriptionService) List(ctx context.Context, userID *uuid.UUID, serviceName *string, page, pageSize int) ([]models.Subscription, int64, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"service_name": serviceName,
		"page":         page,
		"page_size":    pageSize,
	}).Debug("listing subscriptions")

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	subs, total, err := s.repo.List(ctx, userID, serviceName, pageSize, offset)
	if err != nil {
		s.logger.WithError(err).Errorf("failed to list subscriptions")
		return nil, 0, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subs, total, nil
}

func (s *subscriptionService) GetTotalCost(ctx context.Context, req *models.TotalCostRequest) (*models.TotalCostResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id":      req.UserID,
		"service_name": req.ServiceName,
		"start_date":   req.StartDate.Time,
		"end_date":     req.EndDate.Time,
	}).Info("calculating total cost")

	if req.EndDate.Time.Before(req.StartDate.Time) {
		return nil, ErrInvalidDateRange
	}

	periodStart := req.StartDate.Time
    periodEnd := req.EndDate.Time

	subs, err := s.repo.GetActiveSubscriptions(ctx, req.UserID, req.ServiceName, periodStart, periodEnd)
	if err != nil {
		s.logger.WithError(err).Error("failed to calculate total cost")
		return nil, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	totalCost := 0
    for _, sub := range subs {
        months := calculateMonthsInPeriod(sub, periodStart, periodEnd)
        cost := sub.Price * months
        totalCost += cost

        s.logger.WithFields(logrus.Fields{
            "subscription_id": sub.ID,
            "service_name":    sub.ServiceName,
            "price":           sub.Price,
            "months":          months,
            "cost":            cost,
        }).Debug("calculated subscription cost for period")
    }

	s.logger.WithField("total_cost", totalCost).Info("total cost calculated successfully")
	return &models.TotalCostResponse{TotalCost: totalCost}, nil
}

func calculateMonthsInPeriod(sub models.Subscription, periodStart, periodEnd time.Time) int {
    activeStart := sub.StartDate
    if activeStart.Before(periodStart) {
        activeStart = periodStart
    }

    activeEnd := periodEnd
    if sub.EndDate != nil && sub.EndDate.Before(periodEnd) {
        activeEnd = *sub.EndDate
    }

    years := activeEnd.Year() - activeStart.Year()
    months := int(activeEnd.Month()) - int(activeStart.Month())
    return years*12 + months + 1
}