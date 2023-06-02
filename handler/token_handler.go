package handler

import (
	"fmt"
	"log"
	"net/http"
	"newFeatures/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RefreshToken(ctx *gin.Context) {
	var inputTokens *models.GenerateTokens
	if err := ctx.ShouldBindJSON(&inputTokens); err != nil {
		ctx.JSON(http.StatusBadRequest,
			models.ErrorResponseAuth{Message: "Request sent to the server is invalid",
				ResponseError: err.Error(),
				Status:        fmt.Sprintf("%v", http.StatusUnauthorized)})
		return
	}

	tokens, err := h.services.Authorization.RefreshToken(inputTokens.RefreshToken)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized,
			models.ErrorResponseAuth{ResponseError: fmt.Sprintf("loading env: %s", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, tokens)
}
