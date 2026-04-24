package repository

import (
	"REST_service/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, int64, error)
	GetActiveSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate, endDate time.Time) ([]models.Subscription, error)
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Subscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *subscriptionRepository) List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, int64, error) {
	var subs []models.Subscription
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Subscription{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if serviceName != nil {
		query = query.Where("service_name = ?", *serviceName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&subs).Error
	return subs, total, err
}

func (r *subscriptionRepository) GetActiveSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate, endDate time.Time) ([]models.Subscription, error) {
	var subs []models.Subscription
    
    query := r.db.WithContext(ctx).
        Where("start_date <= ?", endDate).
        Where("end_date IS NULL OR end_date >= ?", startDate)
    
    if userID != nil {
        query = query.Where("user_id = ?", *userID)
    }
    if serviceName != nil && *serviceName != "" {
        query = query.Where("service_name = ?", *serviceName)
    }
    
    err := query.Find(&subs).Error
    return subs, err
}