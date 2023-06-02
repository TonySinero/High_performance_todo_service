package service

import (
	"context"
	"errors"
	"newFeatures/models"
	"newFeatures/repository"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TodoPostgresService interface {
	GetTodo(id int) (*models.Todo, error)
	GetTodos(page, limit int64) ([]models.Todo, int, error)
	CreateTodo(todo *models.Todo) (int, error)
	UpdateTodo(todo *models.Todo) error
	DeleteTodoByID(id int) (int, error)
}
type TodoMongoService interface {
	GetTodo(id primitive.ObjectID) (*models.TodoMongo, error)
	GetTodos(page, limit int64) ([]models.TodoMongo, int, error)
	CreateTodo(todo *models.TodoMongo) (string, error)
	UpdateTodo(todo *models.TodoMongo) error
	DeleteTodoByID(id primitive.ObjectID) (string, error)
}
type TodoElasticService interface {
	GetTodo(ctx context.Context, id string) (*models.TodoElastic, error)
	GetTodos(ctx context.Context, page, limit int64) ([]models.TodoElastic, error)
	CreateTodo(ctx context.Context, input *models.TodoElastic) (string, error)
	UpdateTodo(ctx context.Context, todo *models.TodoElastic) (string, error)
	DeleteTodoByID(ctx context.Context, id string) error
	SearchTodos(ctx context.Context, query string, page, limit int64) ([]models.TodoElastic, error)
}
type TodoCassandraService interface {
	CreateTodo(ctx context.Context, todo models.TodoCassandra) error
	UpdateTodo(ctx context.Context, todo models.TodoCassandra) error
	DeleteTodoByID(ctx context.Context, id gocql.UUID) error
	GetTodos(ctx context.Context, page int, limit []byte) ([]models.TodoCassandra, []byte, error)
	GetTodoByID(ctx context.Context, id gocql.UUID) (models.TodoCassandra, error)
}
type TodoMariaService interface {
	CreateTodo(ctx context.Context, todo *models.TodoMaria) (int, error)
	UpdateTodo(ctx context.Context, todo *models.TodoMaria) error
	DeleteTodoByID(ctx context.Context, id int) error
	GetTodos(ctx context.Context, page int64, limit int64) ([]models.TodoMaria, error)
	GetTodoByID(ctx context.Context, id int) (models.TodoMaria, error)
}
type TodoClickHouseService interface {
	CreateTodo(ctx context.Context, todo *models.TodoClickHouse) error
	UpdateTodo(ctx context.Context, todo *models.TodoClickHouse) error
	DeleteTodo(ctx context.Context, id uuid.UUID) error
	GetTodos(ctx context.Context, page, limit int64) ([]models.TodoClickHouse, error)
	GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoClickHouse, error)
}

type TodoCockroachService interface {
	CreateTodo(ctx context.Context, todo *models.TodoCockroach) error
	UpdateTodo(ctx context.Context, todo *models.TodoCockroach) error
	DeleteTodo(ctx context.Context, id uuid.UUID) error
	GetTodos(ctx context.Context, page, limit int) ([]models.TodoCockroach, error)
	GetTodoByID(ctx context.Context, id uuid.UUID) (*models.TodoCockroach, error)
}

type Authorization interface {
	CreateUser(ctx context.Context, user *models.User) (*models.GenerateTokens, error)
	AuthUser(ctx context.Context, user *models.User) (tokens *models.GenerateTokens, err error)
	User(ctx context.Context, userID int) (*models.ResponseUser, error)
	Users(ctx context.Context, page, limit int64) ([]models.ResponseUser, error)
	UpdateUser(ctx context.Context, inputUser *models.ResponseUser) error
	DeleteUser(ctx context.Context, userID int) error
	RefreshToken(refreshToken string) (*models.GenerateTokens, error)
	CheckRole(neededRoles []string, givenRole string) error
	ParseToken(token string) (int, string, error)
	GenerateTokens(user *models.User) (*models.GenerateTokens, error)
	RestorePassword(ctx context.Context, restore *models.RestorePassword) error
}

type Service struct {
	TodoPostgresService
	TodoMongoService
	TodoElasticService
	TodoCassandraService
	TodoMariaService
	TodoClickHouseService
	TodoCockroachService
	Authorization
}

var serviceFactories = map[string]func(*repository.Repository) interface{}{
	repository.PostgresDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoPostgresService: &PostgresService{repository: r},
			Authorization:       &AuthorizationService{repository: r},
		}
	},
	repository.MongoDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoMongoService: &MongoService{repository: r},
		}
	},
	repository.ElasticSearchDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoElasticService: &ElasticService{repository: r},
		}
	},
	repository.CassandraDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoCassandraService: &CassandraService{repository: r},
		}
	},
	repository.MariaDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoMariaService: &MariaService{repository: r},
		}
	},
	repository.ClickHouseDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoClickHouseService: &ClickHouseService{repository: r},
		}
	},
	repository.CockroachDB: func(r *repository.Repository) interface{} {
		return &Service{
			TodoCockroachService: &CockroachService{repository: r},
		}
	},
}

func NewTodoService(dbType string, db *repository.Repository) (*Service, error) {
	serviceFactory, ok := serviceFactories[dbType]
	if !ok {
		return nil, errors.New("unsupported database type")
	}

	return serviceFactory(db).(*Service), nil
}
