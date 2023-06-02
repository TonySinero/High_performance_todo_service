package repository

import (
	"context"
	"fmt"
	"newFeatures/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TodoMongo struct {
	db *mongo.Client
}

func NewTodoMongo(db *mongo.Client) *TodoMongo {
	return &TodoMongo{db: db}
}

func (r *TodoMongo) GetTodoByID(id primitive.ObjectID) (*models.TodoMongo, error) {
	var todo models.TodoMongo
	filter := bson.M{"_id": id}
	collection := r.db.Database("mydb").Collection("todos")
	err := collection.FindOne(context.Background(), filter).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("GetTodoByID: document not found:%w", err)
		}
		return nil, fmt.Errorf("GetTodoByID: repository error:%w", err)
	}
	return &todo, nil
}

func (r *TodoMongo) GetTodos(page, limit int64) ([]models.TodoMongo, int, error) {
	var Todos []models.TodoMongo
	filter := bson.M{}
	collection := r.db.Database("mydb").Collection("todos")
	findOptions := options.Find()
	if page != 0 && limit != 0 {
		findOptions.SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	}
	cur, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTodos: repository error:%w", err)
	}
	for cur.Next(context.Background()) {
		var Todo models.TodoMongo
		err := cur.Decode(&Todo)
		if err != nil {
			return nil, 0, fmt.Errorf("GetTodos: error while decoding todo:%w", err)
		}
		Todos = append(Todos, Todo)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, fmt.Errorf("GetTodos: error during cursor iteration:%w", err)
	}
	cur.Close(context.Background())
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTodos: error while getting count of documents:%w", err)
	}
	return Todos, int(count), nil
}

func (r *TodoMongo) CreateTodo(todo *models.TodoMongo) (string, error) {
	collection := r.db.Database("mydb").Collection("todos")
	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return "", fmt.Errorf("CreateTodo: repository error:%w", err)
	}
	idStr := result.InsertedID.(primitive.ObjectID).Hex()

	if err != nil {
		return "", fmt.Errorf("CreateTodo: error converting id to int: %w", err)
	}
	return idStr, nil
}

func (r *TodoMongo) UpdateTodo(todo *models.TodoMongo) error {
	filter := bson.M{"_id": todo.ID}
	collection := r.db.Database("mydb").Collection("todos")
	update := bson.M{
		"$set": bson.M{
			"title": todo.Title,
			"done":  todo.Done,
		},
	}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("UpdateTodo: repository error:%w", err)
	}
	return nil
}

func (r *TodoMongo) DeleteTodoByID(id primitive.ObjectID) (string, error) {
	filter := bson.M{"_id": id}
	collection := r.db.Database("mydb").Collection("todos")
	var Todo models.TodoMongo
	err := collection.FindOneAndDelete(context.Background(), filter).Decode(&Todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("DeleteTodoByID: document not found")
		}
		return "", fmt.Errorf("DeleteTodoByID: repository error:%w", err)
	}
	return Todo.ID.Hex(), nil
}
