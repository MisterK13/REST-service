package handlers

import (
	"errors"
    "gorm.io/gorm"

	"REST_service/internal/models"
	"REST_service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
	logger  *logrus.Logger
}

func NewSubscriptionHandler(service service.SubscriptionService, logger *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	sub, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("failed to create subscription")

		if errors.Is(err, service.ErrInvalidDateRange) {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
            return
        }

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
        return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription godoc
// @Summary Get subscription by ID
// @Description Get a subscription record by its ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("failed to get subscription")
        
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
        return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Update an existing subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body models.UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	sub, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithError(err).Error("failed to update subscription")
        
        if errors.Is(err, service.ErrSubscriptionNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
            return
        }
        
        if errors.Is(err, service.ErrInvalidDateRange) {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
        return
	}

	c.JSON(http.StatusOK, sub)
}

// DeleteSubscription godoc
// @Summary Delete subscription
// @Description Delete a subscription record
// @Tags subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("failed to delete subscription")
        
        if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, service.ErrSubscriptionNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
        return
	}

	c.Status(http.StatusNoContent)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description Get a paginated list of subscriptions with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} ListResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id UUID format"})
			return
		}
		userID = &id
	}

	var serviceName *string
	if sn := c.Query("service_name"); sn != "" {
		serviceName = &sn
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	subs, total, err := h.service.List(c.Request.Context(), userID, serviceName, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Data:       subs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	})
}

// GetTotalCost godoc
// @Summary Calculate total subscription cost
// @Description Calculate total cost of subscriptions for a given period with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param start_date query string true "Start date (MM-YYYY)"
// @Param end_date query string true "End date (MM-YYYY)"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	var req models.TotalCostRequest

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id UUID format"})
			return
		}
		req.UserID = &id
	}

	if sn := c.Query("service_name"); sn != "" {
		req.ServiceName = &sn
	}

	if err := req.StartDate.UnmarshalParam(c.Query("start_date")); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, use MM-YYYY"})
		return
	}

	if err := req.EndDate.UnmarshalParam(c.Query("end_date")); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_date format, use MM-YYYY"})
		return
	}

	result, err := h.service.GetTotalCost(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("failed to calculate total cost")
        
        if errors.Is(err, service.ErrInvalidDateRange) {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
        return
	}

	c.JSON(http.StatusOK, result)
}

type ErrorResponse struct {
	Error string `json: "error"`
}

type ListResponse struct {
	Data       []models.Subscription `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

func (h *SubscriptionHandler) parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	idStr := c.Param(paramName)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid UUID format"})
		return uuid.Nil, false
	}
	return id, true
}