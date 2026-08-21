package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Venamuss/webhook-dispatcher/internal/subscription"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (repo *PostgresRepository) CreateEndpoint(ctx context.Context, endpoint *subscription.Endpoint) error {
	query := `
		INSERT INTO endpoints (
			id, tenant_id, url, description, secret_key,
			subscribed_events, is_active, rate_limit_rps, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := repo.pool.Exec(ctx, query, endpoint.ID, endpoint.TenantID, endpoint.URL, endpoint.Description, endpoint.SecretKey,
		endpoint.SubscribedEvents, endpoint.IsActive, endpoint.RateLimitRPS, endpoint.CreatedAt, endpoint.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	return nil
}

func (repo *PostgresRepository) GetEndpointByID(ctx context.Context, id string) (*subscription.Endpoint, error) {
	query := `SELECT id, tenant_id, url, description, secret_key, subscribed_events, is_active, rate_limit_rps, created_at, updated_at 
	FROM endpoints 
	WHERE id = $1`

	var result subscription.Endpoint

	err := repo.pool.QueryRow(ctx, query, id).Scan(&result.ID, &result.TenantID, &result.URL, &result.Description, &result.SecretKey,
		&result.SubscribedEvents, &result.IsActive, &result.RateLimitRPS, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, subscription.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get endpoint by id: %w", err)
	}

	return &result, nil
}

func (repo *PostgresRepository) ListEndpointsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]subscription.Endpoint, error) {
	query := `SELECT id, tenant_id, url, description, secret_key, subscribed_events, is_active, rate_limit_rps, created_at, updated_at 
	FROM endpoints 
	WHERE tenant_id = $1 
	ORDER BY created_at DESC 
	LIMIT $2 OFFSET $3`

	rows, err := repo.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}
	defer rows.Close()

	endpoints := make([]subscription.Endpoint, 0, limit)

	for rows.Next() {
		var endpoint subscription.Endpoint
		if err := rows.Scan(&endpoint.ID, &endpoint.TenantID, &endpoint.URL, &endpoint.Description, &endpoint.SecretKey,
			&endpoint.SubscribedEvents, &endpoint.IsActive, &endpoint.RateLimitRPS, &endpoint.CreatedAt, &endpoint.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint: %w", err)
		}
		endpoints = append(endpoints, endpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over endpoints: %w", err)
	}

	return endpoints, nil
}

func (repo *PostgresRepository) SaveDeliveryAttempt(ctx context.Context, attempt *subscription.DeliveryAttempt) error {
	query := `INSERT INTO delivery_attempts (event_id, endpoint_id, attempt_number, status, http_status_code,
		execution_time_ms, request_headers, response_body, error_message, trace_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var httpStatus any = attempt.HTTPStatusCode
	if attempt.HTTPStatusCode == 0 {
		httpStatus = nil
	}

	_, err := repo.pool.Exec(ctx, query, attempt.EventID, attempt.EndpointID, attempt.AttemptNumber, attempt.Status, httpStatus,
		attempt.ExecutionTimeMS, attempt.RequestHeaders, attempt.ResponseBody, attempt.ErrorMessage, attempt.TraceID)
	if err != nil {
		return fmt.Errorf("failed to save delivery attempt: %w", err)
	}

	return nil
}

func (repo *PostgresRepository) UpdateEndpoint(ctx context.Context, id string, endpoint *subscription.Endpoint) error {
	query := `
		UPDATE endpoints
		SET 
			url = COALESCE($2, url),
			subscribed_events = COALESCE($3, subscribed_events),
			is_active = COALESCE($4, is_active),
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := repo.pool.Exec(ctx, query, id, endpoint.URL, endpoint.SubscribedEvents, endpoint.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}

func (repo *PostgresRepository) RotateSecret(ctx context.Context, id string, newSecret string) error {
	query := `UPDATE endpoints SET secret_key = $1 WHERE id = $2`

	_, err := repo.pool.Exec(ctx, query, newSecret, id)
	if err != nil {
		return err
	}

	return nil
}
func (repo *PostgresRepository) DeleteEndpoint(ctx context.Context, id string) error {
	query := `DELETE FROM endpoints WHERE id = $1`

	result, err := repo.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return subscription.ErrNotFound
	}

	return nil
}

func (repo *PostgresRepository) ListLatestDeliveries(ctx context.Context, tenantID, limit int) ([]*subscription.DeliveryAttempt, error) {
	query := `SELECT * FROM delivery_attempts WHERE tenant_id = $1 LIMIT $2 ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list delivery attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*subscription.DeliveryAttempt

	for rows.Next() {
		var attempt subscription.DeliveryAttempt
		if err := rows.Scan(
			&attempt.ID,
			&attempt.EventID,
			&attempt.EndpointID,
			&attempt.AttemptNumber,
			&attempt.Status,
			&attempt.HTTPStatusCode,
			&attempt.ExecutionTimeMS,
			&attempt.RequestHeaders,
			&attempt.ResponseBody,
			&attempt.ErrorMessage,
			&attempt.TraceID,
			&attempt.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan delivery: %w", err)
		}
		attempts = append(attempts, &attempt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}

	return attempts, nil
}

func (repo *PostgresRepository) GetDelivery(ctx context.Context, id int) (*subscription.DeliveryAttempt, error) {
	query := `SELECT * FROM delivery_attempts WHERE id = $1`

	var attempt subscription.DeliveryAttempt

	err := repo.pool.QueryRow(ctx, query, id).Scan(
		&attempt.ID,
		&attempt.EventID,
		&attempt.EndpointID,
		&attempt.AttemptNumber,
		&attempt.HTTPStatusCode,
		&attempt.ExecutionTimeMS,
		&attempt.RequestHeaders,
		&attempt.ResponseBody,
		&attempt.ErrorMessage,
		&attempt.TraceID,
		&attempt.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to query delivery")
	}

	return &attempt, nil
}

func (repo *PostgresRepository) UpdateDeliveryAttempt(ctx context.Context, attempt *subscription.DeliveryAttempt) error {
	query := `UPDATE delivery_attempts 
	        SET status = $1, updated_at = NOW() 
	        WHERE id = $2`

	_, err := repo.pool.Exec(ctx, query, attempt.Status, attempt.ID)
	if err != nil {
		return fmt.Errorf("failed to update delivery attempt: %w", err)
	}

	return nil
}
