package handler

import (
	"encoding/base64"
	"net/http"
	"newFeatures/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

func (h *Handler) createTodoCassandra(ctx *gin.Context) {
	var todo models.TodoCassandra
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.services.TodoCassandraService.CreateTodo(ctx.Request.Context(), todo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Todo created successfully"})
}

func (h *Handler) updateTodoCassandra(ctx *gin.Context) {
	todoID := ctx.Param("id")

	var todo models.TodoCassandra
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := gocql.ParseUUID(todoID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	todo.ID = id

	if err := h.services.TodoCassandraService.UpdateTodo(ctx.Request.Context(), todo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (h *Handler) deleteTodoCassandra(ctx *gin.Context) {
	todoID := ctx.Param("id")

	id, err := gocql.ParseUUID(todoID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	if err := h.services.TodoCassandraService.DeleteTodoByID(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (h *Handler) getTodosCassandra(ctx *gin.Context) {
	pageStr := ctx.Query("page")
	limitStr := ctx.Query("limit")

	page, err := base64.StdEncoding.DecodeString(limitStr)
	if err != nil {
		page = nil
	}

	limit, err := strconv.Atoi(pageStr)
	if err != nil {
		limit = 10
	}

	todos, newPagingState, err := h.services.TodoCassandraService.GetTodos(ctx.Request.Context(), limit, page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := struct {
		Todos         []models.TodoCassandra `json:"todos"`
		NextPageState string                 `json:"next_page_state,omitempty"`
	}{
		Todos:         todos,
		NextPageState: base64.StdEncoding.EncodeToString(newPagingState),
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) getTodoCassandra(ctx *gin.Context) {
	todoID := ctx.Param("id")

	id, err := gocql.ParseUUID(todoID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo ID"})
		return
	}

	todo, err := h.services.TodoCassandraService.GetTodoByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, todo)
}
