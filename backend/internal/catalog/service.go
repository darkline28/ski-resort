package catalog

import (
	"context"
	"errors"
)

type EquipmentRepository interface {
	ListEquipment(ctx context.Context, category *string) ([]Equipment, error)
	GetEquipmentByID(ctx context.Context, id int) (Equipment, error)
	GetEquipmentCategories(ctx context.Context) ([]Category, error)
	GetEquipmentSetByID(ctx context.Context, id int) (EquipmentSet, error)
	GetEquipmentSets(ctx context.Context) ([]EquipmentSet, error)
}

type EquipmentService struct {
	repo EquipmentRepository
}

func NewEquipmentService(repo EquipmentRepository) *EquipmentService {
	return &EquipmentService{repo: repo}
}

func (s *EquipmentService) GetListEquipments(ctx context.Context, categoty *string) ([]Equipment, error) {
	equipments, err := s.repo.ListEquipment(ctx, categoty)
	if err != nil {
		return nil, err
	}
	return equipments, nil
}

var ErrInvalidID = errors.New("invalid id")

func (s *EquipmentService) GetEquipmentByID(ctx context.Context, id int) (Equipment, error) {
	if id <= 0 {
		return Equipment{}, ErrInvalidID
	}
	equipment, err := s.repo.GetEquipmentByID(ctx, id)
	if err != nil {
		return Equipment{}, err
	}

	return equipment, nil
}

func (s *EquipmentService) GetEquipmentCategories(ctx context.Context) ([]Category, error) {
	categories, err := s.repo.GetEquipmentCategories(ctx)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *EquipmentService) GetEquipmentSetByID(ctx context.Context, id int) (EquipmentSet, error) {
	if id <= 0 {
		return EquipmentSet{}, ErrInvalidID
	}
	set, err := s.repo.GetEquipmentSetByID(ctx, id)
	if err != nil {
		return EquipmentSet{}, err
	}

	return set, nil
}

func (s *EquipmentService) GetEquipmentSets(ctx context.Context) ([]EquipmentSet, error) {
	sets, err := s.repo.GetEquipmentSets(ctx)
	if err != nil {
		return nil, err
	}
	return sets, nil
}
