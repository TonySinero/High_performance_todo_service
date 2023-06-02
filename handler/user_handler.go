package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"newFeatures/models"
	"newFeatures/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) createUser(ctx *gin.Context) {
	var user *models.User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			models.ErrorResponse{Message: "Could not binding JSON"})
		return
	}

	token, err := h.services.Authorization.CreateUser(ctx.Request.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrEmptyFields) || errors.Is(err, service.ErrInvalidPassOrName) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest,
				models.ErrorResponseAuth{Message: "Request sent to the server is invalid",
					ResponseError: err.Error(),
					Status:        fmt.Sprintf("%v", http.StatusBadRequest)})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, token)
}

func (h *Handler) authUser(ctx *gin.Context) {
	var user *models.User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{Message: "Could not binding JSON"})
		log.Printf("Failed to process request: create user: %v", err)
		return
	}

	tokens, err := h.services.Authorization.AuthUser(ctx, user)
	if err != nil {
		if err == service.ErrIncorrectCredentials || err == service.ErrDeletedUser {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponseAuth{Message: "Request sent to the server is invalid",
				ResponseError: err.Error(),
				Status:        fmt.Sprintf("%v", http.StatusBadRequest)})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Server cannot process the request"})
		return
	}
	ctx.JSON(http.StatusOK, tokens)
}

func (h *Handler) getUsers(ctx *gin.Context) {
	var page int64 = 1
	var limit int64 = 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		paramPage, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil || paramPage < 0 {
			logrus.Warnf("Invalid page parameter: %s", err)
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid url query"})
			return
		}
		page = paramPage
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		paramLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || paramLimit < 0 {
			logrus.Warnf("Invalid limit parameter: %s", err)
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid url query"})
			return
		}
		limit = paramLimit
	}

	users, err := h.services.Authorization.Users(ctx, page, limit)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (h *Handler) updateUser(ctx *gin.Context) {
	var inputUser models.ResponseUser
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseAuth{
			Message:       "Request sent to the server is invalid",
			ResponseError: err.Error(),
			Status:        fmt.Sprintf("%v", http.StatusBadRequest),
		})
		return
	}

	inputUser.Id = id

	if err := ctx.ShouldBindJSON(&inputUser); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseAuth{
			Message:       "Request sent to the server is invalid",
			ResponseError: err.Error(),
			Status:        fmt.Sprintf("%v", http.StatusBadRequest),
		})
		return
	}

	if err := h.services.Authorization.UpdateUser(ctx, &inputUser); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponseAuth{
				Message:       "Request sent to the server is invalid",
				ResponseError: err.Error(),
				Status:        fmt.Sprintf("%v", http.StatusBadRequest),
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Server cannot process the request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *Handler) deleteUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		logrus.Warnf("Handler deleteTodo (reading param):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid id"})
		return
	}
	err = h.services.Authorization.DeleteUser(ctx, id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (h *Handler) getUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || userID <= 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			models.ErrorResponseAuth{Message: "Request sent to the server is invalid",
				ResponseError: err.Error(),
				Status:        fmt.Sprintf("%v", http.StatusBadRequest)})
		return
	}
	user, err := h.services.Authorization.User(ctx, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *Handler) restorePassword(ctx *gin.Context) {
	var input models.RestorePassword
	if err := ctx.ShouldBindJSON(&input); err != nil {
		logrus.Warnf("Handler restorePassword (binding JSON):%s", err)
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid request"})
		return
	}

	err := h.services.Authorization.RestorePassword(ctx, &input)
	if err != nil {
		if errors.Is(err, service.ErrorEmailDoesNotExist) {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user password updated successfully"})
}
