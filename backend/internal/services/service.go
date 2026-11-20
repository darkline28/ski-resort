package services

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type ServiceRepository interface {
	ListServices(ctx context.Context) ([]Service, error)
	GetServiceByID(ctx context.Context, id int) (Service, error)
}

type ServiceService struct {
	repo ServiceRepository
}

func NewServiceService(repo ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

var ErrServiceNotFound = errors.New("service not found")
var ErrInvalidServiceID = errors.New("invalid service id")

func (s *ServiceService) ListServices(ctx context.Context) ([]Service, error) {
	return s.repo.ListServices(ctx)
}

func (s *ServiceService) GetServiceByID(ctx context.Context, id int) (Service, error) {
	if id <= 0 {
		return Service{}, ErrInvalidServiceID
	}

	service, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Service{}, ErrServiceNotFound
		}
		return Service{}, err
	}

	return service, nil
}
