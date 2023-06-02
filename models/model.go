package models

import (
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type TodoMongo struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title string             `json:"title" bson:"title"`
	Done  bool               `json:"done" bson:"done"`
}
type TodoResponse struct {
	Todo    *TodoMongo        `json:"todo"`
	Message map[string]string `json:"message"`
}

type TodoElastic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoCassandra struct {
	ID        gocql.UUID `json:"id"`
	Title     string     `json:"title"`
	Completed bool       `json:"completed"`
}

type TodoMaria struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoClickHouse struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Title string    `json:"title" db:"title"`
	Done  uint8     `json:"done" db:"done"`
}

type TodoCockroach struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Completed bool      `json:"completed" db:"completed"`
}

type GenerateTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponseAuth struct {
	Message       string `json:"message"`
	ResponseError string `json:"response_error"`
	Status        string `json:"status"`
}

type User struct {
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}
type ResponseUser struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UserRole string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

type Post struct {
	Email    string
	Password string
}

type RestorePassword struct {
	Email    string `json:"email" binding:"required" validate:"email"`
	Password string `json:"password" binding:"required" validate:"password"`
}
