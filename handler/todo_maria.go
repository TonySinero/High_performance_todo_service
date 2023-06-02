package handler

import (
	"encoding/json"
	"net/http"
	"newFeatures/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) getTodoMaria(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("Handler getTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid id"})
		return
	}

	todo, err := h.cache.Get(ctx, strconv.Itoa(id))
	if err != nil {
		logrus.Errorf("Handler getTodo (cache get): %s", err)
	}

	if todo != "" {
		var t models.TodoMaria
		err := json.Unmarshal([]byte(todo), &t)
		if err != nil {
			logrus.Errorf("Handler getTodo (unmarshaling todo): %s", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to get todo"})
			return
		}
		ctx.JSON(http.StatusOK, t)
		return
	}

	t, err := h.services.TodoMariaService.GetTodoByID(ctx, id)
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

	if err := h.cache.Set(ctx, strconv.Itoa(id), string(jsonTodo)); err != nil {
		logrus.Errorf("Handler getTodo (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, t)
}
func (h *Handler) getTodosMaria(ctx *gin.Context) {
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
	todos, err := h.services.TodoMariaService.GetTodos(ctx, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) createTodoMaria(ctx *gin.Context) {
	var input models.TodoMaria
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler createTodo (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	id, err := h.services.TodoMariaService.CreateTodo(ctx, &input)
	if err != nil {
		if err.Error() == "createTodo: error while scanning for user:pq: duplicate key value violates unique constraint" {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "todo with such a model already exists"})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusCreated, id)
}

func (h *Handler) updateTodoMaria(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("Handler updateTodo (reading param): %s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid id"})
		return
	}

	var input models.TodoMaria
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler updateTodo (binding JSON): %s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}
	input.ID = id

	if err := h.services.TodoMariaService.UpdateTodo(ctx, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	jsonTodo, err := json.Marshal(input)
	if err != nil {
		logrus.Errorf("Handler updateTodoMaria (marshaling todo): %s", err)
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to update todo"})
		return
	}

	if err := h.cache.Set(ctx, strconv.Itoa(id), string(jsonTodo)); err != nil {
		logrus.Errorf("Handler updateTodoMaria (cache set): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (h *Handler) deleteTodoMaria(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Warnf("Handler deleteTodo (reading param): %s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid id"})
		return
	}

	if err := h.services.TodoMariaService.DeleteTodoByID(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if err := h.cache.Delete(ctx, strconv.Itoa(id)); err != nil {
		logrus.Errorf("Handler deleteTodoMaria (cache delete): %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
