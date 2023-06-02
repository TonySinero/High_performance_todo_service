package handler

import (
	"newFeatures/cache"
	"newFeatures/graph"
	"newFeatures/graph/generated"
	"newFeatures/graph/middleware"
	"newFeatures/repository"
	"newFeatures/service"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
)

// Handler type replies for handling gin server requests.
type Handler struct {
	services    *service.Service
	cache       *cache.Cache
	reg         *prometheus.Registry
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
	rabbitConn  *amqp.Connection
	rabbitChan  *amqp.Channel
}

// NewHandler function create handler.
func NewHandler(services *service.Service, cache *cache.Cache, reg *prometheus.Registry, kafkaWriter *kafka.Writer, kafkaReader *kafka.Reader, rabbitConn *amqp.Connection, rabbitChan *amqp.Channel) *Handler {
	return &Handler{
		services:    services,
		cache:       cache,
		reg:         reg,
		kafkaWriter: kafkaWriter,
		kafkaReader: kafkaReader,
		rabbitConn:  rabbitConn,
		rabbitChan:  rabbitChan,
	}
}

func (h *Handler) InitRoutes(dbType string) *gin.Engine {
	r := gin.Default()
	metricsMiddleware := NewMetricsMiddleware(h.reg)
	r.Use(h.CorsMiddleware, metricsMiddleware.Metrics)
	r.GET("/metrics", prometheusHandler(h.reg))

	auth := r.Group("/auth")
	{
		auth.POST("/user", h.createUser)
		auth.POST("/login", h.authUser)
		auth.POST("/restore", h.restorePassword)
		auth.POST("/refresh", h.RefreshToken)
	}
	r.GET("/login", h.handleGoogleLogin)
	r.GET("/callback", h.handleGoogleCallback)

	switch dbType {
	case repository.PostgresDB:
		r.Use(h.parseAuthHeader, h.checkRole)
		h.initPostgresRoutes(r)
	case repository.MongoDB:
		r.Use(h.protect)
		h.initMongoRoutes(r)
	case repository.CassandraDB:
		h.initCassandraRoutes(r)
	case repository.MariaDB:
		h.initMariaRoutes(r)
	case repository.ClickHouseDB:
		h.initClickHouseRoutes(r)
	case repository.CockroachDB:
		h.initCockroachRoutes(r)
	case repository.ElasticSearchDB:
		h.initElasticSearchRoutes(r)
	}

	return r
}

func (h *Handler) initPostgresRoutes(r *gin.Engine) {
	r.PUT("/postgres/user/:id", h.updateUser)
	r.DELETE("/postgres/user/:id", h.deleteUser)
	r.GET("/postgres/users", h.getUsers)
	r.GET("/postgres/user/:id", h.getUser)

	r.GET("/postgres/todos", h.getTodosPostgres)
	r.GET("/postgres/todo/:id", h.getTodoPostgres)
	r.POST("/postgres/todo", h.createTodoPostgres)
	r.PUT("/postgres/todo/:id", h.updateTodoPostgres)
	r.DELETE("/postgres/todo/:id", h.deleteTodoPostgres)
	r.POST("/kafka/producer", h.produceKafkaMessages)
	r.POST("/rabbit/producer", h.produceRabbitMessages)

}

func (h *Handler) initMongoRoutes(r *gin.Engine) {
	r.GET("/mongo/todos", h.getTodosMongo)
	r.GET("/mongo/todo/:id", h.getTodoMongo)
	r.POST("/mongo/todo", h.createTodoMongo)
	r.PUT("/mongo/todo/:id", h.updateTodoMongo)
	r.DELETE("/mongo/todo/:id", h.deleteTodoMongo)
	r.GET("/kafka/consumer", h.consumeKafkaMessages)
	r.GET("/rabbit/consumer", h.consumeRabbitMessages)

}

func (h *Handler) initCassandraRoutes(r *gin.Engine) {
	r.GET("/cassandra/todos", h.getTodosCassandra)
	r.GET("/cassandra/todo/:id", h.getTodoCassandra)
	r.POST("/cassandra/todo", h.createTodoCassandra)
	r.PUT("/cassandra/todo/:id", h.updateTodoCassandra)
	r.DELETE("/cassandra/todo/:id", h.deleteTodoCassandra)
}

func (h *Handler) initMariaRoutes(r *gin.Engine) {
	r.GET("/maria/todos", h.getTodosMaria)
	r.GET("/maria/todo/:id", h.getTodoMaria)
	r.POST("/maria/todo", h.createTodoMaria)
	r.PUT("/maria/todo/:id", h.updateTodoMaria)
	r.DELETE("/maria/todo/:id", h.deleteTodoMaria)
}

func (h *Handler) initClickHouseRoutes(r *gin.Engine) {
	r.GET("/clickhouse/todos", h.getTodosClickhouse)
	r.GET("/clickhouse/todo/:id", h.getTodoClickhouse)
	r.POST("/clickhouse/todo", h.createTodoClickhouse)
	r.PUT("/clickhouse/todo/:id", h.updateTodoClickhouse)
	r.DELETE("/clickhouse/todo/:id", h.deleteTodoClickhouse)
}

func (h *Handler) initCockroachRoutes(r *gin.Engine) {
	r.GET("/cockroach/todos", h.getTodosCockroach)
	r.GET("/cockroach/todo/:id", h.getTodoCockroach)
	r.POST("/cockroach/todo", h.createTodoCockroach)
	r.PUT("/cockroach/todo/:id", h.updateTodoCockroach)
	r.DELETE("/cockroach/todo/:id", h.deleteTodoCockroach)
}

func (h *Handler) initElasticSearchRoutes(r *gin.Engine) {
	r.POST("/query", graphqlHandler(h.services, middleware.AuthMiddleware()))
	r.GET("/", playgroundHandler())
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL playground", "/query")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func graphqlHandler(s *service.Service, authMiddleware gin.HandlerFunc) gin.HandlerFunc {
	h := gqlhandler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Serv: s}}))
	return func(ctx *gin.Context) {
		authMiddleware(ctx)
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func prometheusHandler(reg prometheus.Gatherer) gin.HandlerFunc {
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
