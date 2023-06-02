package service

import (
	"fmt"
	"newFeatures/models"
	"newFeatures/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoService struct {
	repository *repository.Repository
}

func (t *MongoService) GetTodo(id primitive.ObjectID) (*models.TodoMongo, error) {
	user, err := t.repository.AppTodoMongo.GetTodoByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (t *MongoService) GetTodos(page, limit int64) ([]models.TodoMongo, int, error) {
	users, pages, err := t.repository.AppTodoMongo.GetTodos(page, limit)
	if err != nil {
		return nil, 0, err
	}
	return users, pages, nil
}

func (t *MongoService) CreateTodo(todo *models.TodoMongo) (string, error) {
	id, err := t.repository.AppTodoMongo.CreateTodo(todo)
	if err != nil {
		return "", fmt.Errorf("something went wrong when creating a user:%w", err)
	}
	return id, nil
}

func (t *MongoService) UpdateTodo(todo *models.TodoMongo) error {
	err := t.repository.AppTodoMongo.UpdateTodo(todo)
	if err != nil {
		return err
	}
	return nil
}

func (t *MongoService) DeleteTodoByID(id primitive.ObjectID) (string, error) {
	Id, err := t.repository.AppTodoMongo.DeleteTodoByID(id)
	if err != nil {
		return "", err
	}
	return Id, nil
}
