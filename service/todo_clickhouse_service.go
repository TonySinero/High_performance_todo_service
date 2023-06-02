package service

import (
	"context"
	"errors"
	"newFeatures/models"
	"newFeatures/repository"

	"github.com/google/uuid"
)

type ClickHouseService struct {
	repository *repository.Repository
}

func (s *ClickHouseService) CreateTodo(ctx context.Context, todo *models.TodoClickHouse) error {
	if todo == nil {
		return errors.New("todo is nil")
	}

	err := s.repository.AppTodoClickHouse.CreateTodo(ctx, todo)
	if err != nil {
		return err
	}

	return nil
}

func (s *ClickHouseService) UpdateTodo(ctx context.Context, todo *models.TodoClickHouse) error {
	if todo == nil {
		return errors.New("todo is nil")
	}

	err := s.repository.AppTodoClickHouse.UpdateTodo(ctx, todo)
	if err != nil {
		return err
	}

	return nil
}

func (s *ClickHouseService) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	err := s.repository.AppTodoClickHouse.DeleteTodo(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ClickHouseService) GetTodos(ctx context.Context, page, limit int64) ([]models.TodoClickHouse, error) {
	if page < 1 || limit < 1 {
		return nil, errors.New("invalid page or limit value")
	}

	users, err := s.repository.AppTodoClickHouse.GetTodos(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *ClickHouseService) GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoClickHouse, error) {
	user, err := s.repository.AppTodoClickHouse.GetTodoByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
