package service

import (
	"context"
	"newFeatures/models"
	"newFeatures/repository"
)

type MariaService struct {
	repository *repository.Repository
}

func (s *MariaService) CreateTodo(ctx context.Context, todo *models.TodoMaria) (int, error) {
	id, err := s.repository.AppTodoMaria.CreateTodo(ctx, todo)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *MariaService) UpdateTodo(ctx context.Context, todo *models.TodoMaria) error {
	err := s.repository.AppTodoMaria.UpdateTodo(ctx, todo)
	if err != nil {
		return err
	}
	return nil
}

func (s *MariaService) DeleteTodoByID(ctx context.Context, id int) error {
	err := s.repository.AppTodoMaria.DeleteTodoByID(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *MariaService) GetTodos(ctx context.Context, page int64, limit int64) ([]models.TodoMaria, error) {
	todos, err := s.repository.AppTodoMaria.GetTodos(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *MariaService) GetTodoByID(ctx context.Context, id int) (models.TodoMaria, error) {
	todo, err := s.repository.AppTodoMaria.GetTodoByID(ctx, id)
	if err != nil {
		return models.TodoMaria{}, err
	}
	return todo, nil
}
