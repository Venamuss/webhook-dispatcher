package subscription

import (
	"errors"
	"time"
)

type Status string

var (
	ErrNilDB    = errors.New("database connection is nil")
	ErrNotFound = errors.New("endpoint not found")
)

const (
	StatusSuccess    Status = "SUCCESS"
	StatusFailed     Status = "FAILED"
	StatusRetrying   Status = "RETRYING"
	StatusDeadLetter Status = "DEAD_LETTER"
)

type Endpoint struct {
	ID               string    `json:"id" db:"id"`
	TenantID         string    `json:"tenant_id" db:"tenant_id"`
	URL              string    `json:"url" db:"url"`
	Description      string    `json:"description,omitempty" db:"description"`
	SecretKey        string    `json:"secret_key,omitempty" db:"secret_key"`
	SubscribedEvents []string  `json:"subscribed_events" db:"subscribed_events"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	RateLimitRPS     int       `json:"rate_limit_rps" db:"rate_limit_rps"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type DeliveryAttempt struct {
	ID              string            `json:"id" db:"id"`
	EventID         string            `json:"event_id" db:"event_id"`
	EndpointID      string            `json:"endpoint_id" db:"endpoint_id"`
	AttemptNumber   int               `json:"attempt_number" db:"attempt_number"`
	Status          Status            `json:"status" db:"status"`
	HTTPStatusCode  int               `json:"http_status_code,omitempty" db:"http_status_code"`
	ExecutionTimeMS int               `json:"execution_time_ms" db:"execution_time_ms"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" db:"request_headers"`
	ResponseBody    string            `json:"response_body,omitempty" db:"response_body"`
	ErrorMessage    string            `json:"error_message,omitempty" db:"error_message"`
	TraceID         string            `json:"trace_id,omitempty" db:"trace_id"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}
