package repository

import (
	"database/sql"
	"fmt"
	"newFeatures/models"

	"github.com/sirupsen/logrus"
)

type TodoPostgres struct {
	db *sql.DB
}

func NewTodoPostgres(db *sql.DB) *TodoPostgres {
	return &TodoPostgres{db: db}
}

func (u TodoPostgres) GetTodoByID(id int) (*models.Todo, error) {
	var todo models.Todo
	result := u.db.QueryRow("SELECT id, title, done FROM todos WHERE id = $1", id)
	if err := result.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
		logrus.Errorf("GetTodoByID: error while scanning for todo:%s", err)
		return nil, fmt.Errorf("GetTodoByID: repository error:%w", err)
	}
	return &todo, nil
}

func (u *TodoPostgres) GetTodos(page, limit int64) ([]models.Todo, int, error) {
	transaction, err := u.db.Begin()
	if err != nil {
		logrus.Errorf("GetTodos: can not starts transaction:%s", err)
		return nil, 0, fmt.Errorf("GetTodos: can not starts transaction:%w", err)
	}
	var Todos []models.Todo
	var query string
	var pages int
	var rows *sql.Rows
	if page == 0 || limit == 0 {
		query = "SELECT id, title, done FROM todos ORDER BY id"
		rows, err = transaction.Query(query)
		if err != nil {
			logrus.Errorf("GetTodos: can not executes a query:%s", err)
			return nil, 0, fmt.Errorf("GetTodos:repository error:%w", err)
		}
		pages = 1
	} else {
		query = "SELECT id, title, done FROM todos ORDER BY id LIMIT $1 OFFSET $2"
		rows, err = transaction.Query(query, limit, (page-1)*limit)
		if err != nil {
			logrus.Errorf("GetTodos: can not executes a query:%s", err)
			return nil, 0, fmt.Errorf("GetTodos:repository error:%w", err)
		}
	}
	for rows.Next() {
		var Todo models.Todo
		if err := rows.Scan(&Todo.ID, &Todo.Title, &Todo.Done); err != nil {
			logrus.Errorf("Error while scanning for todo:%s", err)
			return nil, 0, fmt.Errorf("GetTodos:repository error:%w", err)
		}
		Todos = append(Todos, Todo)
	}
	if pages == 0 {
		query = "SELECT CEILING(COUNT(id)/$1::float) FROM todos"
		row := transaction.QueryRow(query, limit)
		if err := row.Scan(&pages); err != nil {
			logrus.Errorf("Error while scanning for pages:%s", err)
		}
	}
	return Todos, pages, transaction.Commit()
}

func (u *TodoPostgres) CreateTodo(todo *models.Todo) (int, error) {
	var id int
	row := u.db.QueryRow("INSERT INTO todos (title, done) VALUES ($1, $2) RETURNING id", todo.Title, todo.Done)
	if err := row.Scan(&id); err != nil {
		logrus.Errorf("CreateTodo: error while scanning for todo:%s", err)
		return 0, fmt.Errorf("CreateTodo: error while scanning for todo:%w", err)
	}
	return id, nil
}

func (u *TodoPostgres) UpdateTodo(todo *models.Todo) error {
	_, err := u.db.Exec("UPDATE todos SET title = $1, done = $2 WHERE id = $3", todo.Title, todo.Done, todo.ID)
	if err != nil {
		logrus.Errorf("UpdateTodo: error while updating todo:%s", err)
		return fmt.Errorf("UpdateTodo: error while updating todo:%w", err)
	}
	return nil
}

func (u *TodoPostgres) DeleteTodoByID(id int) (int, error) {
	var todoId int
	row := u.db.QueryRow("DELETE FROM todos WHERE id=$1 RETURNING id", id)
	if err := row.Scan(&todoId); err != nil {
		logrus.Errorf("DeleteTodoByID: error while scanning for todoId:%s", err)
		return 0, fmt.Errorf("DeleteTodoByID: error while scanning for todoId:%w", err)
	}
	return todoId, nil
}
