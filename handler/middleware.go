package handler

import (
	"fmt"
	"net/http"
	"newFeatures/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (h *Handler) CorsMiddleware(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "*")
	ctx.Header("Access-Control-Allow-Headers", "*")
	ctx.Header("Content-Type", "application/json")

	if ctx.Request.Method != "OPTIONS" {
		ctx.Next()
	} else {
		ctx.AbortWithStatus(http.StatusOK)
	}
}

func (h *Handler) parseAuthHeader(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Message: "empty auth header"})
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Message: "invalid header"})
		return
	}

	if len(headerParts[1]) == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Message: "token is empty"})
		return
	}

	id, role, err := h.services.Authorization.ParseToken(headerParts[1])

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Message: err.Error()})
		return
	}
	ctx.Set("role", role)
	ctx.Set("id", id)
}

func (h *Handler) checkRole(ctx *gin.Context) {
	necessaryRole := []string{string(models.RoleUser), string(models.RoleAdmin)}
	if err := h.services.Authorization.CheckRole(necessaryRole, ctx.GetString("role")); err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Message: "not enough rights"})
		return
	}
}

type MetricsMiddleware struct {
	OpsProcessed          *prometheus.CounterVec
	ReqDuration           *prometheus.HistogramVec
	TaskCreated           prometheus.Counter
	TaskCompleted         prometheus.Counter
	TaskOperationDuration prometheus.Histogram
	TaskErrors            prometheus.Counter
	ActiveTasks           prometheus.Gauge
}

func NewMetricsMiddleware(reg prometheus.Registerer) *MetricsMiddleware {
	m := &MetricsMiddleware{
		OpsProcessed: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "todo_service_processed_ops_total",
			Help: "The total number of processed events",
		}, []string{"method", "path", "status_code"}),
		ReqDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "todo_service_request_duration_seconds",
			Help:    "http_server_request_duration_seconds",
			Buckets: []float64{.01, .05, .1, .5, 1, 5, 10, 15},
		}, []string{"method", "path", "status_code"}),
		TaskCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "todo_service_tasks_created_total",
			Help: "The total number of created tasks",
		}),
		TaskCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "todo_service_tasks_completed_total",
			Help: "The total number of completed tasks",
		}),
		TaskOperationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "todo_service_task_operation_duration_seconds",
			Help:    "The duration of task operations",
			Buckets: []float64{0.1, 0.5, 1, 5},
		}),
		TaskErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "todo_service_task_errors_total",
			Help: "The total number of task errors",
		}),
	}
	reg.MustRegister(
		m.OpsProcessed,
		m.ReqDuration,
		m.TaskCreated,
		m.TaskCompleted,
		m.TaskOperationDuration,
		m.TaskErrors,
	)
	return m
}

func (lm *MetricsMiddleware) Metrics(ctx *gin.Context) {
	start := time.Now()
	ctx.Next()
	lm.OpsProcessed.With(prometheus.Labels{
		"method":      ctx.Request.Method,
		"path":        fmt.Sprintf("%v", ctx.Request.URL),
		"status_code": fmt.Sprintf("%v", ctx.Writer.Status()),
	}).Inc()
	lm.ReqDuration.WithLabelValues(
		ctx.Request.Method,
		fmt.Sprintf("%v", ctx.Request.URL),
		fmt.Sprintf("%v", ctx.Writer.Status()),
	).Observe(time.Since(start).Seconds())
	lm.TaskCreated.Inc()
	lm.TaskCompleted.Inc()
	lm.TaskOperationDuration.Observe(time.Since(start).Seconds())
	lm.TaskErrors.Inc()
}
