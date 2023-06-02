package main

import (
	"context"
	"database/sql"
	"fmt"
	"newFeatures/cache"
	"newFeatures/database"
	"newFeatures/handler"
	"newFeatures/repository"
	"newFeatures/server"
	"newFeatures/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Fatalf("Error loading .env file. %s", err.Error())
	}

	port := os.Getenv("API_SERVER_PORT")
	currentDB := os.Getenv("CURRENT_DB")
	kafkaWriter, kafkaReader, err := initializeKafka()
	if err != nil {
		logrus.Fatalf("Failed to initialize Kafka: %v", err)
	}
	conn, channel, err := initializeRabbitMQ()
	if err != nil {
		logrus.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	cache := initializeRedis()
	r, dbType, err := getRepository(currentDB)
	if err != nil {
		logrus.Fatalf("Error occurred while initializing the repository: %s", err.Error())
	}

	s, err := service.NewTodoService(dbType, r)
	if err != nil {
		return
	}
	reg := prometheus.NewRegistry()
	handler := handler.NewHandler(s, cache, reg, kafkaWriter, kafkaReader, conn, channel)
	routes := handler.InitRoutes(dbType)

	server := new(server.Server)
	go func() {
		if err := server.Run(port, routes); err != nil {
			logrus.Fatalf("Error occurred while running HTTPS server: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error occurred while shutting down HTTPS server: %s", err.Error())
	}
}

func getRepository(currentDB string) (*repository.Repository, string, error) {
	var dbType string
	var db interface{}
	var err error

	switch currentDB {
	case "postgres":
		dbType = repository.PostgresDB
		db, err = initializePostgresDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize Postgres database: %s", err.Error())
		}
	case "mongo":
		dbType = repository.MongoDB
		db, err = initializeMongoDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize MongoDB: %s", err.Error())
		}
	case "elasticsearch":
		dbType = repository.ElasticSearchDB
		db, err = initializeElasticSearch()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize ElasticSearch: %s", err.Error())
		}
	case "cassandra":
		dbType = repository.CassandraDB
		db, err = initializeCassandraDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize Cassandra: %s", err.Error())
		}
	case "maria":
		dbType = repository.MariaDB
		db, err = initializeMariaDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize MariaDB: %s", err.Error())
		}
	case "clickhouse":
		dbType = repository.ClickHouseDB
		db, err = initializeClickHouseDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize ClickHouseDB: %s", err.Error())
		}
	case "cockroach":
		dbType = repository.CockroachDB
		db, err = initializeCockroachDB()
		if err != nil {
			return nil, "", fmt.Errorf("failed to initialize CockroachDB: %s", err.Error())
		}
	default:
		return nil, "", fmt.Errorf("unsupported database type: %s", currentDB)
	}
	repo, err := repository.NewRepository(dbType, db)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize repository: %s", err.Error())
	}
	return repo, dbType, nil
}

func initializePostgresDB() (*sql.DB, error) {
	if os.Getenv("POSTGRES_HOST") == "" || os.Getenv("POSTGRES_PORT") == "" || os.Getenv("POSTGRES_USER") == "" || os.Getenv("POSTGRES_PASSWORD") == "" || os.Getenv("POSTGRES_DB") == "" || os.Getenv("POSTGRES_SSL_MODE") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.NewPostgresDB(database.PostgresDB{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	})
}

func initializeMongoDB() (*mongo.Client, error) {
	if os.Getenv("MONGO_HOST") == "" || os.Getenv("MONGO_PORT") == "" || os.Getenv("MONGO_USERNAME") == "" || os.Getenv("MONGO_PASSWORD") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.ConnectToMongo(database.MongoDB{
		Host:     os.Getenv("MONGO_HOST"),
		Port:     os.Getenv("MONGO_PORT"),
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})
}

func initializeRedis() *cache.Cache {
	return cache.NewCache(
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PASSWORD"),
		0,
		10*time.Minute,
	)
}
func initializeElasticSearch() (*elasticsearch.Client, error) {
	if os.Getenv("ELASTIC_HOST") == "" || os.Getenv("ELASTIC_USERNAME") == "" || os.Getenv("ELASTIC_PASSWORD") == "" || os.Getenv("ELASTIC_INDEX") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.NewElasticSearchDB(database.ElasticSearchDB{
		Host:     os.Getenv("ELASTIC_HOST"),
		Username: os.Getenv("ELASTIC_USERNAME"),
		Password: os.Getenv("ELASTIC_PASSWORD"),
		Index:    os.Getenv("ELASTIC_INDEX"),
	})
}

func initializeCassandraDB() (*gocql.Session, error) {
	if os.Getenv("CASSANDRA_HOST") == "" || os.Getenv("CASSANDRA_KEYSPACE") == "" || os.Getenv("CASSANDRA_USERNAME") == "" || os.Getenv("CASSANDRA_PASSWORD") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.ConnectToCassandra(database.CassandraDB{
		Host:     os.Getenv("CASSANDRA_HOST"),
		Keyspace: os.Getenv("CASSANDRA_KEYSPACE"),
		Username: os.Getenv("CASSANDRA_USERNAME"),
		Password: os.Getenv("CASSANDRA_PASSWORD"),
	})
}

func initializeMariaDB() (*sql.DB, error) {
	if os.Getenv("MARIA_USER") == "" || os.Getenv("MARIA_PASSWORD") == "" || os.Getenv("MARIA_HOST") == "" || os.Getenv("MARIA_PORT") == "" || os.Getenv("MARIA_DB") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.NewMariaDB(database.MariaDB{
		Host:     os.Getenv("MARIA_HOST"),
		Port:     os.Getenv("MARIA_PORT"),
		Username: os.Getenv("MARIA_USER"),
		Password: os.Getenv("MARIA_PASSWORD"),
		DBName:   os.Getenv("MARIA_DB"),
	})
}

func initializeClickHouseDB() (*sql.DB, error) {
	if os.Getenv("CLICKHOUSE_HOST") == "" || os.Getenv("CLICKHOUSE_PORT") == "" || os.Getenv("CLICKHOUSE_USER") == "" || os.Getenv("CLICKHOUSE_PASSWORD") == "" || os.Getenv("CLICKHOUSE_DB") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.NewClickHouseDB(database.ClickHouseDB{
		Host:     os.Getenv("CLICKHOUSE_HOST"),
		Port:     os.Getenv("CLICKHOUSE_PORT"),
		Username: os.Getenv("CLICKHOUSE_USER"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		DBName:   os.Getenv("CLICKHOUSE_DB"),
	})
}

func initializeCockroachDB() (*sql.DB, error) {
	if os.Getenv("COCKROACH_HOST") == "" || os.Getenv("COCKROACH_PORT") == "" || os.Getenv("COCKROACH_USERNAME") == "" || os.Getenv("COCKROACH_PASSWORD") == "" || os.Getenv("COCKROACH_DB") == "" {
		return nil, fmt.Errorf("some of the required environment variables are not set")
	}

	return database.NewCockroachDB(database.CockroachDB{
		Host:     os.Getenv("COCKROACH_HOST"),
		Port:     os.Getenv("COCKROACH_PORT"),
		Username: os.Getenv("COCKROACH_USERNAME"),
		Password: os.Getenv("COCKROACH_PASSWORD"),
		DBName:   os.Getenv("COCKROACH_DB"),
	})
}

func initializeKafka() (*kafka.Writer, *kafka.Reader, error) {
	brokers := []string{"localhost:9092"}
	topic := "todo"

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  "test-consumer-group",
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return kafkaWriter, kafkaReader, nil
}
func initializeRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://username:password@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	err = channel.ExchangeDeclare(
		"todo_exchange",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	_, err = channel.QueueDeclare(
		"todo_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	err = channel.QueueBind(
		"todo_queue",
		"todo.key1",
		"todo_exchange",
		false,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	return conn, channel, nil
}
