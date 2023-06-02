package repository

import (
	"context"
	"errors"
	"newFeatures/models"

	"github.com/gocql/gocql"
)

type TodoCassandra struct {
	session *gocql.Session
}

func NewTodoCassandraDB(session *gocql.Session) *TodoCassandra {
	return &TodoCassandra{session: session}
}

func (r *TodoCassandra) CreateTodo(ctx context.Context, todo models.TodoCassandra) error {
	todo.ID = gocql.TimeUUID()

	query := r.session.Query(`
		INSERT INTO todos (id, title, completed) VALUES (?, ?, ?)
	`, todo.ID, todo.Title, todo.Completed).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return err
	}

	return nil
}

func (r *TodoCassandra) UpdateTodo(ctx context.Context, todo models.TodoCassandra) error {
	query := r.session.Query(`
		UPDATE todos SET title = ?, completed = ? WHERE id = ? IF EXISTS
	`, todo.Title, todo.Completed, todo.ID).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return err
	}

	return nil
}

func (r *TodoCassandra) DeleteTodoByID(ctx context.Context, id gocql.UUID) error {
	query := r.session.Query(`
		DELETE FROM todos WHERE id = ?
	`, id).WithContext(ctx)

	if err := query.Exec(); err != nil {
		return err
	}

	return nil
}

func (r *TodoCassandra) GetTodos(ctx context.Context, page int, limit []byte) ([]models.TodoCassandra, []byte, error) {
	query := r.session.Query("SELECT id, title, completed FROM todos").WithContext(ctx)

	query.PageSize(page)
	query.PageState(limit)

	iter := query.Iter()
	defer iter.Close()

	todos := make([]models.TodoCassandra, 0)
	var id gocql.UUID
	var title string
	var completed bool

	for iter.Scan(&id, &title, &completed) {
		todos = append(todos, models.TodoCassandra{
			ID:        id,
			Title:     title,
			Completed: completed,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, nil, err
	}

	newPagingState := iter.PageState()

	return todos, newPagingState, nil
}

func (r *TodoCassandra) GetTodoByID(ctx context.Context, id gocql.UUID) (models.TodoCassandra, error) {
	var todo models.TodoCassandra
	if err := r.session.Query(`
		SELECT id, title, completed FROM todos WHERE id = ?
	`, id).WithContext(ctx).Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
		if err == gocql.ErrNotFound {
			return models.TodoCassandra{}, errors.New("todo not found")
		}
		return models.TodoCassandra{}, err
	}

	return todo, nil
}
