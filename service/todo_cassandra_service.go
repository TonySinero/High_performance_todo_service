package service

import (
	"context"
	"errors"
	"fmt"
	"newFeatures/models"
	"newFeatures/repository"

	"github.com/gocql/gocql"
)

type CassandraService struct {
	repository *repository.Repository
}

func (s *CassandraService) CreateTodo(ctx context.Context, todo models.TodoCassandra) error {
	err := s.repository.AppTodoCassandra.CreateTodo(ctx, todo)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

func (s *CassandraService) UpdateTodo(ctx context.Context, todo models.TodoCassandra) error {
	err := s.repository.AppTodoCassandra.UpdateTodo(ctx, todo)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	return nil
}

func (s *CassandraService) DeleteTodoByID(ctx context.Context, id gocql.UUID) error {
	err := s.repository.AppTodoCassandra.DeleteTodoByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

func (s *CassandraService) GetTodos(ctx context.Context, page int, limit []byte) ([]models.TodoCassandra, []byte, error) {
	todos, newPagingState, err := s.repository.AppTodoCassandra.GetTodos(ctx, page, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return todos, newPagingState, nil
}

func (s *CassandraService) GetTodoByID(ctx context.Context, id gocql.UUID) (models.TodoCassandra, error) {
	todo, err := s.repository.AppTodoCassandra.GetTodoByID(ctx, id)
	if err != nil {
		if err.Error() == "todo not found" {
			return models.TodoCassandra{}, errors.New("todo not found")
		}
		return models.TodoCassandra{}, fmt.Errorf("failed to get todo by ID: %w", err)
	}

	return todo, nil
}
