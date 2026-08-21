package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Venamuss/webhook-dispatcher/internal/subscription"
)

type DeliveryHandler struct {
	service *subscription.Service
}

func NewDeliveryHandler(service *subscription.Service) *DeliveryHandler {
	return &DeliveryHandler{
		service: service,
	}
}

// GET	/api/v1/deliveries/{id}
func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.GetDelivery(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("service failed: %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delivery)
}

// POST /api/v1/deliveries/{id}/retry
func (h *DeliveryHandler) RetryDelivery(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Delivery ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "failed to parse delivery ID", http.StatusInternalServerError)
		return
	}

	// Fetch the delivery attempt
	// delivery, err := h.service.GetDelivery(r.Context(), id)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("failed to retrieve delivery: %v", err), http.StatusInternalServerError)
	// 	return
	// }

	// Retry the delivery
	err = h.service.RetryDelivery(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retry delivery: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Delivery retried successfully",
	})
}
