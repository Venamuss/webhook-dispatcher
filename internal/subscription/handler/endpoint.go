package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Venamuss/webhook-dispatcher/internal/platform/httputil"
	"github.com/Venamuss/webhook-dispatcher/internal/subscription"
)

type EndpointHandler struct {
	service *subscription.Service
}

func NewEndpointHandler(service *subscription.Service) *EndpointHandler {
	return &EndpointHandler{
		service: service,
	}
}

// POST	/api/v1/endpoints
func (h *EndpointHandler) NewEndpoint(w http.ResponseWriter, r *http.Request) {
	var endpoint subscription.Endpoint

	err := httputil.DecodeJSON(w, r, &endpoint)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = h.service.CreateEndpoint(r.Context(), &endpoint)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)

	//TODO Ответ сделать от POST /api/v1/endpoints
	w.Write([]byte("Endpoint created"))
}

// GET	/api/v1/endpoints
func (h *EndpointHandler) GetEndpointsWithFilter(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if tenantID == "" {
		http.Error(w, "tenant_id is required", 400)
		return
	}
	endpoints, err := h.service.ListEndpointsByTenant(r.Context(), tenantID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(endpoints)
}

// GET	/api/v1/endpoints/{id}
func (h *EndpointHandler) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	endpoint, err := h.service.GetEndpointByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(endpoint)
}

// PATCH	/api/v1/endpoints/{id}
func (h *EndpointHandler) UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var endpoint subscription.Endpoint
	err := httputil.DecodeJSON(w, r, &endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the endpoint with the provided data
	err = h.service.UpdateEndpoint(r.Context(), id, &endpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Endpoint updated"})
}

// DELETE	/api/v1/endpoints/{id}
func (h *EndpointHandler) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteEndpoint(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Endpoint deleted"})
}

// POST /api/v1/endpoints/{id}/rotate-secret
func (h *EndpointHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request struct {
		OldSecret string `json:"old_secret"`
	}

	err := httputil.DecodeJSON(w, r, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.RotateSecret(r.Context(), id, request.OldSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Secret rotated"})

}

// GET	/api/v1/endpoints/{id}/deliveries
func (h *EndpointHandler) ListLatestDeliveries(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Endpoint ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "failed to parse endpoint ID", http.StatusInternalServerError)
	}
	limit := 0

	limitStr := r.URL.Query().Get("limit")
	if strings.EqualFold(limitStr, "") {
		limit = 10
	}
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if limit < 1 {
		limit = 1
	} else if limit > 100 {
		limit = 100
	}

	deliveries, err := h.service.ListLatestDeliveries(r.Context(), id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deliveries)
}
