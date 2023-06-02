package handler

import (
	"encoding/json"
	"net/http"
	"newFeatures/models"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"strconv"
)

func (h *Handler) getTodoPostgres(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("Handler getTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid id"})
		return
	}

	cacheKey := strconv.Itoa(id)
	todo, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		logrus.Errorf("Handler getTodo (cache get): %s", err)
	}

	if todo != "" {
		var t models.Todo
		err := json.Unmarshal([]byte(todo), &t)
		if err != nil {
			logrus.Errorf("Handler getTodo (unmarshaling todo): %s", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to get todo"})
			return
		}
		ctx.JSON(http.StatusOK, t)
		return
	}

	t, err := h.services.TodoPostgresService.GetTodo(id)
	if err != nil {
		logrus.Errorf("Handler getTodo (db get): %s", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to get todo"})
		return
	}

	jsonTodo, err := json.Marshal(t)
	if err != nil {
		logrus.Errorf("Handler getTodo (marshaling todo): %s", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to get todo"})
		return
	}

	if err := h.cache.Set(ctx, cacheKey, string(jsonTodo)); err != nil {
		logrus.Errorf("Handler getTodo (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) getTodosPostgres(ctx *gin.Context) {
	var page int64 = 1
	var limit int64 = 10

	if ctx.Query("page") != "" {
		paramPage, err := strconv.ParseInt(ctx.Query("page"), 10, 64)
		if err != nil || paramPage < 0 {
			logrus.Warnf("No url request:%s", err)
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid url query"})
			return
		}
		page = paramPage
	}
	if ctx.Query("limit") != "" {
		paramLimit, err := strconv.ParseInt(ctx.Query("limit"), 10, 64)
		if err != nil || paramLimit < 0 {
			logrus.Warnf("No url request:%s", err)
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid url query"})
			return
		}
		limit = paramLimit
	}
	todos, pages, err := h.services.TodoPostgresService.GetTodos(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	ctx.Header("pages", strconv.Itoa(pages))
	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) createTodoPostgres(ctx *gin.Context) {
	var input models.Todo
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler createTodo (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	id, err := h.services.TodoPostgresService.CreateTodo(&input)
	if err != nil {
		if err.Error() == "createTodo: error while scanning for user:pq: duplicate key value violates unique constraint" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "todo with such an model already exists"})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusCreated, id)
}

func (h *Handler) updateTodoPostgres(ctx *gin.Context) {
	var input models.Todo
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		logrus.Warnf("Handler getTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler updateTodo (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}
	input.ID = id
	err = h.services.TodoPostgresService.UpdateTodo(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	cacheKey := strconv.Itoa(id)
	jsonTodo, err := json.Marshal(input)
	if err != nil {
		logrus.Errorf("Handler updateTodo (marshaling todo): %s", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to update todo"})
		return
	}

	if err := h.cache.Set(ctx, cacheKey, string(jsonTodo)); err != nil {
		logrus.Errorf("Handler updateTodo (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (h *Handler) deleteTodoPostgres(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		logrus.Warnf("Handler deleteTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid id"})
		return
	}

	cacheKey := strconv.Itoa(id)
	_, err = h.services.TodoPostgresService.DeleteTodoByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		logrus.Errorf("Handler deleteTodo (cache delete): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (h *Handler) produceKafkaMessages(ctx *gin.Context) {
	var input models.Todo
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("binding JSON: %s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		logrus.Errorf("failed to marshal JSON: %v", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	msg := kafka.Message{
		Value: jsonData,
	}

	err = h.kafkaWriter.WriteMessages(ctx, msg)
	if err != nil {
		logrus.Errorf("failed to produce message: %v", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo sent successfully"})
}

func (h *Handler) produceRabbitMessages(ctx *gin.Context) {
	var input models.Todo
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	err = h.rabbitChan.Publish(
		"todo_exchange",
		"todo.key1",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo sent successfully"})
}
