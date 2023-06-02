package repository

import (
	"context"
	"database/sql"
	"fmt"
	"newFeatures/models"

	"github.com/google/uuid"
)

type TodoCockroach struct {
	DB *sql.DB
}

func NewTodoCockroachDB(db *sql.DB) *TodoCockroach {
	return &TodoCockroach{DB: db}
}

func (r *TodoCockroach) GetTodos(ctx context.Context, page, limit int) ([]models.TodoCockroach, error) {
	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, title, completed FROM todos LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []models.TodoCockroach{}
	for rows.Next() {
		var todo models.TodoCockroach
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (r *TodoCockroach) GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoCockroach, error) {
	var todo models.TodoCockroach
	err := r.DB.QueryRowContext(ctx, "SELECT id, title, completed FROM todos WHERE id = $1", id).Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *TodoCockroach) CreateTodo(ctx context.Context, todo *models.TodoCockroach) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO todos (title, completed) VALUES ($1, $2)", todo.Title, todo.Completed)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TodoCockroach) UpdateTodo(ctx context.Context, todo *models.TodoCockroach) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE todos SET title = $1, completed = $2 WHERE id = $3", todo.Title, todo.Completed, todo.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TodoCockroach) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
