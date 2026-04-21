package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceName string     `gorm:"not null" json:"service_name"`
	Price       int        `gorm:"not null" json:"price"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartDate   time.Time  `gorm:"not null" json:"start_date" example:"01-2026"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"12-2026"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string      `json:"service_name" binding:"required"`
	Price       int         `json:"price" binding:"required,min=1"`
	UserID      uuid.UUID   `json:"user_id" binding:"required"`
	StartDate   CustomTime  `json:"start_date" binding:"required" swaggertype:"string" example:"04-2026"`
	EndDate     *CustomTime `json:"end_date,omitempty" swaggertype:"string" example:"12-2026"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string     `json:"service_name,omitempty"`
	Price       *int        `json:"price,omitempty" binding:"omitempty,min=1"`
	StartDate   *CustomTime `json:"start_date,omitempty" swaggertype:"string" example:"04-2026"`
	EndDate     *CustomTime `json:"end_date,omitempty" swaggertype:"string" example:"12-2026"`
}

type TotalCostRequest struct {
	UserID      *uuid.UUID `form:"user_id"`
	ServiceName *string    `form:"service_name"`
	StartDate   CustomTime `form:"start_date" binding:"required"`
	EndDate     CustomTime `form:"end_date" binding:"required"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

func (s *Subscription) MarshalJSON() ([]byte, error) {
	type Alias Subscription
	aux := &struct {
		StartDate string  `json:"start_date"`
		EndDate   *string `json:"end_date,omitempty"`
		*Alias
	}{
		Alias:     (*Alias)(s),
		StartDate: s.StartDate.Format("01-2006"),
	}

	if s.EndDate != nil {
		endDateStr := s.EndDate.Format("01-2006")
		aux.EndDate = &endDateStr
	}

	return json.Marshal(aux)
}
