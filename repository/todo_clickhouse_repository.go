package repository

import (
	"context"
	"database/sql"
	"fmt"
	"newFeatures/models"

	"github.com/google/uuid"
)

type TodoClickHouse struct {
	DB *sql.DB
}

func NewTodoClickHouseDB(db *sql.DB) *TodoClickHouse {
	return &TodoClickHouse{DB: db}
}

func (r *TodoClickHouse) GetTodos(ctx context.Context, page, limit int64) ([]models.TodoClickHouse, error) {
	offset := (page - 1) * limit

	query := fmt.Sprintf("SELECT id, title, done FROM todos LIMIT %d, %d", offset, limit)
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.TodoClickHouse
	for rows.Next() {
		var todo models.TodoClickHouse
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Done)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *TodoClickHouse) GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoClickHouse, error) {
	var todo models.TodoClickHouse
	err := r.DB.QueryRowContext(ctx, "SELECT id, title, done FROM todos WHERE id = ?", id).Scan(&todo.ID, &todo.Title, &todo.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("todo not found")
		}
		return nil, err
	}
	return &todo, nil
}

func (r *TodoClickHouse) CreateTodo(ctx context.Context, todo *models.TodoClickHouse) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO todos (title, done) VALUES (?, ?)", todo.Title, todo.Done)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (r *TodoClickHouse) UpdateTodo(ctx context.Context, todo *models.TodoClickHouse) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "ALTER TABLE todos UPDATE title = ?, done = ? WHERE id = ?", todo.Title, todo.Done, todo.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *TodoClickHouse) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "ALTER TABLE todos DELETE WHERE id = ?", id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
