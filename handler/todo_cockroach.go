package handler

import (
	"net/http"
	"newFeatures/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) getTodosCockroach(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	todos, err := h.services.TodoCockroachService.GetTodos(ctx, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get todos" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) getTodoCockroach(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	todo, err := h.services.TodoCockroachService.GetTodoByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, todo)
}

func (h *Handler) createTodoCockroach(ctx *gin.Context) {
	var todo models.TodoCockroach
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.services.TodoCockroachService.CreateTodo(ctx, &todo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Todo created successfully"})
}

func (h *Handler) updateTodoCockroach(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	var todo models.TodoCockroach
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	todo.ID = id
	err = h.services.TodoCockroachService.UpdateTodo(ctx, &todo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (h *Handler) deleteTodoCockroach(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	err = h.services.TodoCockroachService.DeleteTodo(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
