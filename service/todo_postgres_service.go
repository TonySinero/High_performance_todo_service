package service

import (
	"fmt"
	"newFeatures/models"
	"newFeatures/repository"
)

type PostgresService struct {
	repository *repository.Repository
}

func (t *PostgresService) GetTodo(id int) (*models.Todo, error) {
	user, err := t.repository.AppTodoPostgres.GetTodoByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (t *PostgresService) GetTodos(page, limit int64) ([]models.Todo, int, error) {
	users, pages, err := t.repository.AppTodoPostgres.GetTodos(page, limit)
	if err != nil {
		return nil, 0, err
	}
	return users, pages, nil
}

func (t *PostgresService) CreateTodo(todo *models.Todo) (int, error) {
	id, err := t.repository.AppTodoPostgres.CreateTodo(todo)
	if err != nil {
		return 0, fmt.Errorf("something went wrong when creating a user:%w", err)
	}
	return id, nil
}

func (t *PostgresService) UpdateTodo(todo *models.Todo) error {
	err := t.repository.AppTodoPostgres.UpdateTodo(todo)
	if err != nil {
		return err
	}
	return nil
}

func (t *PostgresService) DeleteTodoByID(id int) (int, error) {
	Id, err := t.repository.AppTodoPostgres.DeleteTodoByID(id)
	if err != nil {
		return 0, err
	}
	return Id, nil
}
