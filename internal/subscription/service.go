package subscription

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateEndpoint(ctx context.Context, endpoint Endpoint) error
	GetEndpointByID(ctx context.Context, id string) (Endpoint, error)
	ListEndpointsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]Endpoint, error)
	SaveDeliveryAttempt(ctx context.Context, attempt DeliveryAttempt) error
	UpdateEndpoint(ctx context.Context, id string, endpoint *Endpoint) error
	DeleteEndpoint(ctx context.Context, id string) error
	RotateSecret(ctx context.Context, id string, newSecret string) error
	ListLatestDeliveries(ctx context.Context, tenantID, limit int) ([]*DeliveryAttempt, error)
	GetDelivery(ctx context.Context, id int) (*DeliveryAttempt, error)
	UpdateDeliveryAttempt(ctx context.Context, attempt *DeliveryAttempt) error
}
type Config struct {
	DefaultRateLimit int
}

type Service struct {
	repo Repository
	cfg  Config
}

func NewService(repo Repository, cfg Config) *Service {
	if cfg.DefaultRateLimit == 0 {
		cfg.DefaultRateLimit = 50
	}
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *Service) RecordDeliveryAttempt(ctx context.Context, attempt *DeliveryAttempt) error {
	return s.repo.SaveDeliveryAttempt(ctx, *attempt)
}

func GenerateSecretKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("whsec_%x", bytes), nil
}

// TODO НЕ ЗАБЫТЬ В ДОКЕРЕ ДОБАВИТЬ КЛЮЧ И RATE В ENV
func (s *Service) CreateEndpoint(ctx context.Context, endpoint *Endpoint) error {
	parsedURL, err := url.ParseRequestURI(endpoint.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	endpoint.URL = parsedURL.String()
	endpoint.ID = uuid.NewString()
	secret, err := GenerateSecretKey()
	if err != nil {
		return err
	}
	endpoint.SecretKey = secret

	now := time.Now().UTC()
	endpoint.CreatedAt = now
	endpoint.UpdatedAt = now
	endpoint.IsActive = true

	if endpoint.RateLimitRPS <= 0 {
		endpoint.RateLimitRPS = s.cfg.DefaultRateLimit
	}

	return s.repo.CreateEndpoint(ctx, *endpoint)
}

func (s *Service) ListEndpointsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]Endpoint, error) {
	return s.repo.ListEndpointsByTenant(ctx, tenantID, limit, offset)
}

func (s *Service) GetEndpointByID(ctx context.Context, endpointID string) (*Endpoint, error) {
	endpoint, err := s.repo.GetEndpointByID(ctx, endpointID)
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (s *Service) UpdateEndpoint(ctx context.Context, id string, endpoint *Endpoint) error {
	existingEndpoint, err := s.repo.GetEndpointByID(ctx, id)
	if err != nil {
		return err
	}

	existingEndpoint.URL = endpoint.URL
	existingEndpoint.SubscribedEvents = endpoint.SubscribedEvents
	existingEndpoint.IsActive = endpoint.IsActive
	existingEndpoint.UpdatedAt = time.Now().UTC()

	return s.repo.UpdateEndpoint(ctx, id, &existingEndpoint)
}

func (s *Service) DeleteEndpoint(ctx context.Context, id string) error {
	err := s.repo.DeleteEndpoint(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	return nil
}

func (s *Service) RotateSecret(ctx context.Context, id string, oldSecret string) error {
	endpoint, err := s.repo.GetEndpointByID(ctx, id)
	if err != nil {
		return fmt.Errorf("repository failed: %w", err)
	}

	if endpoint.SecretKey != oldSecret {
		return fmt.Errorf("invalid secret key")
	}

	newSecret, err := GenerateSecretKey()
	if err != nil {
		return fmt.Errorf("failed to generate secret key: %w", err)
	}

	err = s.repo.RotateSecret(ctx, id, newSecret)
	if err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}

func (s *Service) ListLatestDeliveries(ctx context.Context, tenantID, limit int) ([]*DeliveryAttempt, error) {
	deliveries, err := s.repo.ListLatestDeliveries(ctx, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("repository failed: %w", err)
	}
	return deliveries, nil
}

func (s *Service) GetDelivery(ctx context.Context, id int) (*DeliveryAttempt, error) {
	return s.repo.GetDelivery(ctx, id)
}

func (s *Service) RetryDelivery(ctx context.Context, id int) error {
	// Fetch the delivery attempt
	delivery, err := s.repo.GetDelivery(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	delivery.Status = "pending"
	delivery.UpdatedAt = time.Now().UTC()

	return s.repo.UpdateDeliveryAttempt(ctx, delivery)
}
