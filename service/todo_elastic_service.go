package service

import (
	"context"
	"fmt"
	"newFeatures/models"
	"newFeatures/repository"
)

type ElasticService struct {
	repository *repository.Repository
}

func (s *ElasticService) CreateTodo(ctx context.Context, todo *models.TodoElastic) (string, error) {
	// Call ElasticSearch's CreateTodo function
	id, err := s.repository.AppTodoElasticSearch.CreateTodo(ctx, todo)
	if err != nil {
		return "", fmt.Errorf("failed to create todo: %w", err)
	}
	return id, nil
}

func (s *ElasticService) GetTodo(ctx context.Context, id string) (*models.TodoElastic, error) {
	// Call ElasticSearch's GetTodoByID function
	todoElastic, err := s.repository.AppTodoElasticSearch.GetTodoByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Perform any additional operations on todo here

	return todoElastic, nil
}

func (s *ElasticService) GetTodos(ctx context.Context, page, limit int64) ([]models.TodoElastic, error) {
	// Call ElasticSearch's GetTodos function
	todoElastics, err := s.repository.AppTodoElasticSearch.GetTodos(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}
	// Perform any additional operations on todos here

	return todoElastics, nil
}

func (s *ElasticService) UpdateTodo(ctx context.Context, todo *models.TodoElastic) (string, error) {

	// Call ElasticSearch's UpdateTodo function
	id, err := s.repository.AppTodoElasticSearch.UpdateTodo(ctx, todo)
	if err != nil {
		return "", fmt.Errorf("failed to update todo: %w", err)
	}
	return id, nil
}

func (s *ElasticService) DeleteTodoByID(ctx context.Context, id string) error {
	// Call ElasticSearch's DeleteTodoByID function
	err := s.repository.AppTodoElasticSearch.DeleteTodoByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}

func (s *ElasticService) SearchTodos(ctx context.Context, query string, page, limit int64) ([]models.TodoElastic, error) {
	// Call ElasticSearch's SearchTodos function
	todos, err := s.repository.AppTodoElasticSearch.SearchTodos(ctx, query, page, limit)
	if err != nil {
		return nil, err
	}
	return todos, nil
}
