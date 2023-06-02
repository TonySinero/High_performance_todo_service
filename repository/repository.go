package repository

import (
	"context"
	"database/sql"
	"errors"
	"newFeatures/models"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AppTodoPostgres interface {
	GetTodoByID(id int) (*models.Todo, error)
	GetTodos(page, limit int64) ([]models.Todo, int, error)
	CreateTodo(todo *models.Todo) (int, error)
	UpdateTodo(todo *models.Todo) error
	DeleteTodoByID(id int) (int, error)
}
type AppTodoMongo interface {
	GetTodoByID(id primitive.ObjectID) (*models.TodoMongo, error)
	GetTodos(page, limit int64) ([]models.TodoMongo, int, error)
	CreateTodo(todo *models.TodoMongo) (string, error)
	UpdateTodo(todo *models.TodoMongo) error
	DeleteTodoByID(id primitive.ObjectID) (string, error)
}
type AppTodoElasticSearch interface {
	GetTodoByID(ctx context.Context, id string) (*models.TodoElastic, error)
	GetTodos(ctx context.Context, page, limit int64) ([]models.TodoElastic, error)
	CreateTodo(ctx context.Context, input *models.TodoElastic) (string, error)
	UpdateTodo(ctx context.Context, todo *models.TodoElastic) (string, error)
	DeleteTodoByID(ctx context.Context, id string) error
	SearchTodos(ctx context.Context, query string, page, limit int64) ([]models.TodoElastic, error)
}
type AppTodoCassandra interface {
	CreateTodo(ctx context.Context, todo models.TodoCassandra) error
	UpdateTodo(ctx context.Context, todo models.TodoCassandra) error
	DeleteTodoByID(ctx context.Context, id gocql.UUID) error
	GetTodos(ctx context.Context, page int, limit []byte) ([]models.TodoCassandra, []byte, error)
	GetTodoByID(ctx context.Context, id gocql.UUID) (models.TodoCassandra, error)
}

type AppTodoMaria interface {
	CreateTodo(ctx context.Context, todo *models.TodoMaria) (int, error)
	UpdateTodo(ctx context.Context, todo *models.TodoMaria) error
	DeleteTodoByID(ctx context.Context, id int) error
	GetTodos(ctx context.Context, page int64, limit int64) ([]models.TodoMaria, error)
	GetTodoByID(ctx context.Context, id int) (models.TodoMaria, error)
}

type AppTodoClickHouse interface {
	CreateTodo(ctx context.Context, todo *models.TodoClickHouse) error
	UpdateTodo(ctx context.Context, todo *models.TodoClickHouse) error
	DeleteTodo(ctx context.Context, id uuid.UUID) error
	GetTodos(ctx context.Context, page, limit int64) ([]models.TodoClickHouse, error)
	GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoClickHouse, error)
}

type AppTodoCockroach interface {
	CreateTodo(ctx context.Context, todo *models.TodoCockroach) error
	UpdateTodo(ctx context.Context, todo *models.TodoCockroach) error
	DeleteTodo(ctx context.Context, id uuid.UUID) error
	GetTodos(ctx context.Context, page, limit int) ([]models.TodoCockroach, error)
	GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoCockroach, error)
}

type AuthorizationApp interface {
	CreateUser(ctx context.Context, user *models.User) error
	CheckByEmail(ctx context.Context, restore *models.RestorePassword) error
	UserById(ctx context.Context, userID int) (*models.ResponseUser, error)
	UserByPhone(ctx context.Context, user *models.User) (*models.User, error)
	Users(ctx context.Context, page, limit int64) ([]models.ResponseUser, error)
	UpdateUser(ctx context.Context, inputUser *models.ResponseUser) error
	DeleteUser(ctx context.Context, userID int) error
	UserRoleById(userId int) (*models.User, error)
	RestorePassword(ctx context.Context, restore *models.RestorePassword) error
}

const (
	PostgresDB      string = "postgres"
	MongoDB         string = "mongo"
	ElasticSearchDB string = "elasticsearch"
	CassandraDB     string = "cassandra"
	MariaDB         string = "maria"
	ClickHouseDB    string = "clickhouse"
	CockroachDB     string = "cockroach"
)

type Repository struct {
	AppTodoPostgres
	AppTodoMongo
	AppTodoElasticSearch
	AppTodoCassandra
	AppTodoMaria
	AppTodoClickHouse
	AppTodoCockroach
	AuthorizationApp
}

func NewRepository(dbType string, db interface{}) (*Repository, error) {
	switch dbType {
	case "postgres":
		PostgresDB, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database postgres connection")
		}
		return &Repository{
			AppTodoPostgres:  NewTodoPostgres(PostgresDB),
			AuthorizationApp: NewAuthRepository(PostgresDB),
		}, nil
	case "mongo":
		MongoDB, ok := db.(*mongo.Client)
		if !ok {
			return nil, errors.New("invalid database mongo connection")
		}
		return &Repository{
			AppTodoMongo: NewTodoMongo(MongoDB),
		}, nil
	case "elasticsearch":
		ElasticSearchDB, ok := db.(*elasticsearch.Client)
		if !ok {
			return nil, errors.New("invalid database elasticsearch connection")
		}
		return &Repository{
			AppTodoElasticSearch: NewTodoElasticSearch(ElasticSearchDB, os.Getenv("ELASTIC_INDEX")),
		}, nil
	case "cassandra":
		CassandraDB, ok := db.(*gocql.Session)
		if !ok {
			return nil, errors.New("invalid database cassandra connection")
		}
		return &Repository{
			AppTodoCassandra: NewTodoCassandraDB(CassandraDB),
		}, nil
	case "maria":
		MariaDB, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database maria connection")
		}
		return &Repository{
			AppTodoMaria: NewTodoMaria(MariaDB),
		}, nil
	case "clickhouse":
		ClickHouseDB, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database clickhouse connection")
		}
		return &Repository{
			AppTodoClickHouse: NewTodoClickHouseDB(ClickHouseDB),
		}, nil
	case "cockroach":
		CockroachDB, ok := db.(*sql.DB)
		if !ok {
			return nil, errors.New("invalid database cockroach connection")
		}
		return &Repository{
			AppTodoCockroach: NewTodoCockroachDB(CockroachDB),
		}, nil
	default:
		return nil, errors.New("unsupported database type")
	}
}
