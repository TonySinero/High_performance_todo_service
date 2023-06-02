package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"newFeatures/models"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Google OAuth2 endpoint
	oauth2Endpoint = google.Endpoint
	// Google OAuth2 configuration
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: oauth2Endpoint,
	}
)

func (h *Handler) handleGoogleLogin(ctx *gin.Context) {
	url := oauth2Config.AuthCodeURL(generateState())
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) handleGoogleCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "message": err.Error()})
		return
	}

	accessToken := token.AccessToken
	ctx.JSON(http.StatusOK, models.GenerateTokens{
		AccessToken: accessToken,
	})
}

func (h *Handler) protect(ctx *gin.Context) {
	token, err := getTokenFromHeader(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}
	if userInfo.Name != "" {
		ctx.Set("name", userInfo.Name)
	}
	if userInfo.Email != "" {
		ctx.Set("email", userInfo.Email)
	}
}

func getTokenFromHeader(ctx *gin.Context) (*oauth2.Token, error) {
	tokenStr := extractTokenFromAuthHeader(ctx)
	if tokenStr == "" {
		return nil, errors.New("token not found")
	}

	token := &oauth2.Token{
		AccessToken: tokenStr,
	}
	return token, nil
}

func extractTokenFromAuthHeader(ctx *gin.Context) string {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		return ""
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return ""
	}

	return headerParts[1]
}

func generateState() string {
	stateBytes := make([]byte, 32)
	_, err := rand.Read(stateBytes)
	if err != nil {
		return ""
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)
	return state
}
