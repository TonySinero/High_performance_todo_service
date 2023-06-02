package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"newFeatures/models"
	"newFeatures/repository"

	"github.com/google/uuid"
)

type CockroachService struct {
	repository *repository.Repository
}

func (s *CockroachService) GetTodos(ctx context.Context, page, limit int) ([]models.TodoCockroach, error) {
	todos, err := s.repository.AppTodoCockroach.GetTodos(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}
	return todos, nil
}

func (s *CockroachService) GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoCockroach, error) {
	todo, err := s.repository.AppTodoCockroach.GetTodoByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("todo not found")
		}
		return nil, fmt.Errorf("failed to get todo by ID: %w", err)
	}
	return todo, nil
}

func (s *CockroachService) CreateTodo(ctx context.Context, todo *models.TodoCockroach) error {
	err := s.repository.AppTodoCockroach.CreateTodo(ctx, todo)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}
	return nil
}

func (s *CockroachService) UpdateTodo(ctx context.Context, todo *models.TodoCockroach) error {
	err := s.repository.AppTodoCockroach.UpdateTodo(ctx, todo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("todo not found")
		}
		return fmt.Errorf("failed to update todo: %w", err)
	}
	return nil
}

func (s *CockroachService) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	err := s.repository.AppTodoCockroach.DeleteTodo(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("todo not found")
		}
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}
