package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"net/http"
	"newFeatures/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *Handler) getTodoMongo(ctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("type conversion error:%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid id"})
		return
	}

	cacheKey := fmt.Sprintf("todo:%s", id.Hex())
	jsonTodo, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		logrus.Errorf("getTodoMongo (cache get): %s", err)
	}

	if jsonTodo != "" {
		var todo models.TodoMongo
		if err := json.Unmarshal([]byte(jsonTodo), &todo); err != nil {
			logrus.Errorf("getTodoMongo (unmarshaling todo): %s", err)
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to get todo"})
			return
		}
		ctx.JSON(http.StatusOK, todo)
		return
	}

	todo, err := h.services.TodoMongoService.GetTodo(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Todo not found"})
			return
		}
		logrus.Errorf("getTodoMongo (db get): %s", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to get todo"})
		return
	}

	byteTodo, err := json.Marshal(todo)
	if err != nil {
		logrus.Errorf("getTodoMongo (marshaling todo): %s", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to get todo"})
		return
	}
	jsonTodo = string(byteTodo)

	if err := h.cache.Set(ctx, cacheKey, jsonTodo); err != nil {
		logrus.Errorf("getTodoMongo (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, todo)

}

func (h *Handler) getTodosMongo(ctx *gin.Context) {
	var page int64 = 1
	var limit int64 = 10

	if ctx.Query("page") != "" {
		paramPage, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
		if err != nil || paramPage < 1 {
			logrus.Warnf("No url request:%s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid url query"})
			return
		}
		page = paramPage
	}

	if ctx.Query("limit") != "" {
		paramLimit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
		if err != nil || paramLimit < 1 {
			logrus.Warnf("No url request:%s", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid url query"})
			return
		}
		limit = paramLimit
	}

	todos, pages, err := h.services.TodoMongoService.GetTodos(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to get todos"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"todos": todos,
		"pages": pages,
	})
}

func (h *Handler) createTodoMongo(ctx *gin.Context) {
	var input models.TodoMongo
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler createMongoTodo (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	id, err := h.services.TodoMongoService.CreateTodo(&input)
	if err != nil {
		if we, ok := err.(mongo.WriteException); ok {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "todo with such an model already exists"})
					return
				}
			}
		}
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, id)
}

func (h *Handler) updateTodoMongo(ctx *gin.Context) {
	var input models.TodoMongo
	id := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler updateTodo (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.Warnf("type conversion error:%s", err)
		return
	}
	input.ID = objID
	err = h.services.TodoMongoService.UpdateTodo(&input)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Todo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update todo"})
		return
	}

	cacheKey := fmt.Sprintf("todo:%s", id)
	byteTodo, err := json.Marshal(input)
	if err != nil {
		logrus.Errorf("updateTodoMongo (marshaling todo): %s", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update todo"})
		return
	}
	jsonTodo := string(byteTodo)

	if err := h.cache.Set(ctx, cacheKey, jsonTodo); err != nil {
		logrus.Errorf("updateTodoMongo (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (h *Handler) deleteTodoMongo(ctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("Handler deleteTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid id"})
		return
	}

	ID, delErr := h.services.TodoMongoService.DeleteTodoByID(id)
	if delErr != nil {
		if errors.Is(delErr, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Todo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to delete todo"})
		return
	}

	cacheKey := fmt.Sprintf("todo:%s", ID)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		logrus.Errorf("deleteTodoMongo (cache delete): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (h *Handler) consumeKafkaMessages(ctx *gin.Context) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg, err := h.kafkaReader.ReadMessage(timeoutCtx)
	if err != nil {
		if err == io.EOF {
			ctx.JSON(http.StatusOK, gin.H{"message": "No messages in Kafka queue"})
			return
		}
		logrus.Errorf("failed to read message: %v", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	var todo models.Todo
	if err := json.Unmarshal(msg.Value, &todo); err != nil {
		logrus.Errorf("failed to unmarshal message: %v", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": todo})
}

func (h *Handler) consumeRabbitMessages(ctx *gin.Context) {
	messages, err := h.rabbitChan.Consume(
		"todo_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			ctx.JSON(http.StatusOK, gin.H{"message": "No messages in RabbitMQ queue"})
			return
		case message := <-messages:
			var todo models.Todo
			err := json.Unmarshal(message.Body, &todo)
			if err != nil {
				continue
			}
			ctx.JSON(http.StatusOK, gin.H{"message": todo})
			return
		}
	}
}
