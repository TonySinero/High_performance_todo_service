package repository

import (
	"context"
	"database/sql"
	"newFeatures/models"
)

type TodoMaria struct {
	DB *sql.DB
}

func NewTodoMaria(db *sql.DB) *TodoMaria {
	return &TodoMaria{DB: db}
}

func (r *TodoMaria) CreateTodo(ctx context.Context, todo *models.TodoMaria) (int, error) {
	result, err := r.DB.ExecContext(ctx, "INSERT INTO todos (title, completed) VALUES (?, ?)", todo.Title, todo.Completed)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *TodoMaria) UpdateTodo(ctx context.Context, todo *models.TodoMaria) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE todos SET title = ?, completed = ? WHERE id = ?", todo.Title, todo.Completed, todo.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *TodoMaria) DeleteTodoByID(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, "DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TodoMaria) GetTodos(ctx context.Context, page int64, limit int64) ([]models.TodoMaria, error) {
	offset := (page - 1) * limit
	rows, err := r.DB.QueryContext(ctx, "SELECT id, title, completed FROM todos LIMIT ?, ?", offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []models.TodoMaria{}
	for rows.Next() {
		var todo models.TodoMaria
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (r *TodoMaria) GetTodoByID(ctx context.Context, id int) (models.TodoMaria, error) {
	row := r.DB.QueryRowContext(ctx, "SELECT id, title, completed FROM todos WHERE id = ?", id)
	todo := models.TodoMaria{}
	err := row.Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {
		return models.TodoMaria{}, err
	}
	return todo, nil
}
